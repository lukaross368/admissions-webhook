package webhooks

import (
	"strings"

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

		owners := pod.ObjectMeta.OwnerReferences
		for _, owner := range owners {
			if owner.Kind == "Job" || owner.Kind == "CronJob" {
				return false
			}
		}

		annotations := pod.ObjectMeta.Annotations
		_, ok := annotations["mutation.webhook.admission.k8s.io/skip"]
		if ok {
			if strings.ToLower(annotations["mutation.webhook.admission.k8s.io/skip"]) == "true" {
				return false
			}
		}

		labels := pod.ObjectMeta.Labels
		_, ok = labels["environment"]
		if !ok {
			return false
		}

		if strings.ToLower(labels["environment"]) == "dev" {
			return false
		}

		if strings.ToLower(labels["environment"]) == "test" || strings.ToLower(labels["environment"]) == "preprod" || strings.ToLower(labels["environment"]) == "prod" {
			alwaysRestartPodPolicy := corev1.RestartPolicyAlways
			return pod.Spec.RestartPolicy != alwaysRestartPodPolicy
		}

		return false
	}

	return applyPodPatch(ar, shouldPatchPod, podLabelsPatch)
}
