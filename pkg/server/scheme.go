package server

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"

	admissionv1 "k8s.io/api/admission/v1"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	corev1 "k8s.io/api/core/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

// Scheme and Codecs are shared across the webhook server.
var Scheme = runtime.NewScheme()
var Codecs = serializer.NewCodecFactory(Scheme)

/*
 * This function was originally copied from the Kubernetes project (https://github.com/kubernetes/kubernetes)
 * and is licensed under Apache License 2.0.
 * Modifications have been made by the author of this project.
 */
func init() {
	utilruntime.Must(corev1.AddToScheme(Scheme))
	utilruntime.Must(admissionv1beta1.AddToScheme(Scheme))
	utilruntime.Must(admissionregistrationv1beta1.AddToScheme(Scheme))
	utilruntime.Must(admissionv1.AddToScheme(Scheme))
	utilruntime.Must(admissionregistrationv1.AddToScheme(Scheme))
}
