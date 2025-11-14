package webhooks

import (
	v1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/klog/v2"
)

var (
	scheme = runtime.NewScheme()
	codecs = serializer.NewCodecFactory(scheme)
)

// decodePodFromAdmissionReview handles common logic for decoding a Pod from an AdmissionReview
func decodePodFromAdmissionReview(ar v1.AdmissionReview) (*corev1.Pod, *v1.AdmissionResponse) {
	podResource := metav1.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}
	if ar.Request.Resource != podResource {
		klog.Errorf("expect resource to be %s", podResource)
		return nil, nil
	}

	raw := ar.Request.Object.Raw
	pod := &corev1.Pod{}
	deserializer := codecs.UniversalDeserializer()
	if _, _, err := deserializer.Decode(raw, nil, pod); err != nil {
		klog.Error(err)
		return nil, &v1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	}

	return pod, nil
}

// applyPodPatch mutates a pod if shouldPatchPod returns true
func applyPodPatch(ar v1.AdmissionReview, shouldPatchPod func(*corev1.Pod) bool, patch string) *v1.AdmissionResponse {
	klog.V(2).Info("mutating pods")
	pod, errResp := decodePodFromAdmissionReview(ar)
	if errResp != nil {
		return errResp
	}

	reviewResponse := &v1.AdmissionResponse{
		Allowed: true,
	}

	if shouldPatchPod(pod) {
		reviewResponse.Patch = []byte(patch)
		pt := v1.PatchTypeJSONPatch
		reviewResponse.PatchType = &pt
	}

	return reviewResponse
}

// applyPodValidation validates a pod and returns an AdmissionResponse
func applyPodValidation(ar v1.AdmissionReview, validate func(*corev1.Pod) (bool, *metav1.Status)) *v1.AdmissionResponse {
	klog.V(2).Info("validating pods")
	pod, errResp := decodePodFromAdmissionReview(ar)
	if errResp != nil {
		return errResp
	}

	allowed, status := validate(pod)
	return &v1.AdmissionResponse{
		Allowed: allowed,
		Result:  status,
	}
}
