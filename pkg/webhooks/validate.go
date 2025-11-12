package webhooks

import (
	"encoding/json"
	"fmt"

	v1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

type validationResponse struct {
	Container string `json:"container"`
	Issue     string `json:"issue"`
}

// validate pods function. contains logic for how to validate pods
func ValidatePods(ar v1.AdmissionReview) *v1.AdmissionResponse {

	validate := func(pod *corev1.Pod) (bool, string) {

		if len(pod.Spec.Containers) == 0 {
			return true, "passed validation"
		}

		invalidContainers := []validationResponse{}

		for _, container := range pod.Spec.Containers {
			if container.LivenessProbe == nil {
				invalidContainer := validationResponse{
					Container: container.Name,
					Issue:     fmt.Sprintf("container %s failed validation due to missing Liveness Probe", container.Name),
				}
				invalidContainers = append(invalidContainers, invalidContainer)
			}
			if container.ReadinessProbe == nil {
				invalidContainer := validationResponse{
					Container: container.Name,
					Issue:     fmt.Sprintf("container %s failed validation due to missing Readiness Probe", container.Name),
				}
				invalidContainers = append(invalidContainers, invalidContainer)
			}
		}

		if len(invalidContainers) == 0 {
			return true, "passed validation"
		}

		invalidContainerJson, err := json.Marshal(invalidContainers)
		if err != nil {
			klog.Errorf("Error occurred during marshalling: %s", err.Error())
			return false, "failed validation due to json marshalling error"
		}

		return false, string(invalidContainerJson)
	}
	return applyPodValidation(ar, validate)
}
