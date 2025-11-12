package webhooks

import (
	v1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
)

// mutate pods function. contains logic for how to mutate pods
func MutatePods(ar v1.AdmissionReview) *v1.AdmissionResponse {

	// pod mutation patch
	const podLabelsPatch = `[
		{"op":"add","path":"/metadata/labels/automatedLabel","value":"applied-via-mutating-webhook"}
	]`

	// mutation patch condition function
	shouldPatchPod := func(pod *corev1.Pod) bool {
		labels := pod.ObjectMeta.Labels
		_, ok := labels["automatedLabel"]
		return !ok
	}

	return applyPodPatch(ar, shouldPatchPod, podLabelsPatch)
}
