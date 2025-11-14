package server

import (
	"github.com/lukaross368/admissions-webhook/pkg/webhooks"

	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
)

type admitv1Func func(v1.AdmissionReview) *v1.AdmissionResponse

type admitHandler struct {
	v1 admitv1Func
}

var (
	port     int
	certFile string
	keyFile  string
)

var CmdWebhook = &cobra.Command{
	Use:   "webhook",
	Short: "Start admission webhook server",
	Args:  cobra.MaximumNArgs(0),
	RunE:  runWebhook,
}

func Execute() {
	if err := CmdWebhook.Execute(); err != nil {
		panic(err)
	}
}

func init() {
	CmdWebhook.Flags().StringVar(&certFile, "tls-cert-file", "", "File containing the default x509 Certificate for HTTPS. (CA cert, if any, concatenated after server cert).")
	CmdWebhook.Flags().StringVar(&keyFile, "tls-private-key-file", "", "File containing the default x509 private key matching --tls-cert-file.")
	CmdWebhook.Flags().IntVar(&port, "port", 443, "Secure port that the webhook listens on")
}

func serve(w http.ResponseWriter, r *http.Request, admit admitHandler) {
	var body []byte
	if r.Body != nil {

		if data, err := io.ReadAll(r.Body); err == nil {
			body = data
		}
	}

	// verify the content type is accurate
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		klog.Errorf("contentType=%s, expect application/json", contentType)
		return
	}

	klog.V(2).Info(fmt.Sprintf("handling request: %s", body))

	deserializer := Codecs.UniversalDeserializer()
	obj, gvk, err := deserializer.Decode(body, nil, nil)
	if err != nil {
		msg := fmt.Sprintf("Request could not be decoded: %v", err)
		klog.Error(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	var responseObj runtime.Object
	switch *gvk {
	case v1.SchemeGroupVersion.WithKind("AdmissionReview"):
		requestedAdmissionReview, ok := obj.(*v1.AdmissionReview)
		if !ok {
			klog.Errorf("Expected v1.AdmissionReview but got: %T", obj)
			return
		}
		responseAdmissionReview := &v1.AdmissionReview{}
		responseAdmissionReview.SetGroupVersionKind(*gvk)
		responseAdmissionReview.Response = admit.v1(*requestedAdmissionReview)
		responseAdmissionReview.Response.UID = requestedAdmissionReview.Request.UID
		responseObj = responseAdmissionReview
	default:
		msg := fmt.Sprintf("Unsupported group version kind: %v", gvk)
		klog.Error(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	klog.V(2).Info(fmt.Sprintf("sending response: %v", responseObj))
	respBytes, err := json.Marshal(responseObj)
	if err != nil {
		klog.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(respBytes); err != nil {
		klog.Error(err)
	}
}

func newDelegateToV1AdmitHandler(f admitv1Func) admitHandler {
	return admitHandler{
		v1: f,
	}
}

func serveValidatePods(w http.ResponseWriter, r *http.Request) {
	serve(w, r, newDelegateToV1AdmitHandler(webhooks.ValidatePods))
}

func serveMutatePods(w http.ResponseWriter, r *http.Request) {
	serve(w, r, newDelegateToV1AdmitHandler(webhooks.MutatePods))
}

func runWebhook(cmd *cobra.Command, args []string) error {
	config := Config{
		CertFile: certFile,
		KeyFile:  keyFile,
	}

	muxWebhook := http.NewServeMux()
	muxWebhook.HandleFunc("/mutate-pods", serveMutatePods)
	muxWebhook.HandleFunc("/validate-pods", serveValidatePods)

	go func() {
		muxMetrics := http.NewServeMux()
		muxMetrics.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":2112", muxMetrics)
	}()

	server := &http.Server{
		Addr:      fmt.Sprintf(":%d", port),
		TLSConfig: configTLS(config),
		Handler:   muxWebhook,
	}

	return server.ListenAndServeTLS("", "")
}
