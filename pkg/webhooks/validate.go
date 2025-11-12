package webhooks

import (
	v1 "k8s.io/api/admission/v1"
)

// validate pods function. contains logic for how to validate pods
func ValidatePods(ar v1.AdmissionReview) *v1.AdmissionResponse {
	// validate pods logic

	return &v1.AdmissionResponse{}
}
