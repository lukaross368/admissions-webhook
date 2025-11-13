package webhooks

import (
	"fmt"
	"net/http"
	"strings"

	v1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// validate pods function. contains logic for how to validate pods
func ValidatePods(ar v1.AdmissionReview) *v1.AdmissionResponse {

	validate := func(pod *corev1.Pod) (bool, *metav1.Status) {
		responseStatus := metav1.Status{}

		if len(pod.Spec.Containers) == 0 {
			return true, &responseStatus
		}

		var invalidContainers []string
		for _, container := range pod.Spec.Containers {
			if container.LivenessProbe == nil {
				message := fmt.Sprintf("container %s failed validation due to missing Liveness Probe", container.Name)
				invalidContainers = append(invalidContainers, message)
			}
			if container.ReadinessProbe == nil {
				message := fmt.Sprintf("container %s failed validation due to missing Readiness Probe", container.Name)
				invalidContainers = append(invalidContainers, message)
			}
		}

		if len(invalidContainers) == 0 {
			return true, &responseStatus
		}

		responseStatus.Message = strings.Join(invalidContainers, ", ")
		responseStatus.Code = http.StatusForbidden
		return false, &responseStatus
	}
	return applyPodValidation(ar, validate)
}
