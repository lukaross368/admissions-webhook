package webhooks

import (
	"testing"

	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// helper to create AdmissionReview from raw pod JSON
func newAdmissionReview(podJSON []byte) v1.AdmissionReview {
	return v1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "admission.k8s.io/v1",
			Kind:       "AdmissionReview",
		},
		Request: &v1.AdmissionRequest{
			UID: "123",
			Resource: metav1.GroupVersionResource{
				Group:    "",
				Version:  "v1",
				Resource: "pods",
			},
			Object: runtime.RawExtension{
				Raw: podJSON,
			},
		},
	}
}

// helper to run a mutate pod test
func runMutatePodTest(t *testing.T, podJSON string, expectedPatch string, expectAllowed bool) {
	ar := newAdmissionReview([]byte(podJSON))
	resp := MutatePods(ar)

	if expectedPatch != "" {
		require.JSONEq(t, expectedPatch, string(resp.Patch))
	} else {
		require.Nil(t, resp.Patch)
	}
	require.Equal(t, expectAllowed, resp.Allowed)
}

func TestMutatePods(t *testing.T) {
	patchJSON := `[{"op": "replace", "path": "/spec/restartPolicy", "value": "Always"}]`

	tests := []struct {
		name          string
		podJSON       string
		expectedPatch string
		expectAllowed bool
	}{
		{
			name: "patch pod test env",
			podJSON: `{
				"metadata": {"name": "test-pod", "labels": {"environment": "test"}},
				"spec": {"containers":[{"name":"nginx","image":"nginx:latest"}],"restartPolicy":"Never"}
			}`,
			expectedPatch: patchJSON,
			expectAllowed: true,
		},
		{
			name: "patch pod prod env",
			podJSON: `{
				"metadata": {"name": "test-pod", "labels": {"environment": "PROD"}},
				"spec": {"containers":[{"name":"nginx","image":"nginx:latest"}],"restartPolicy":"Never"}
			}`,
			expectedPatch: patchJSON,
			expectAllowed: true,
		},
		{
			name: "patch pod preprod env",
			podJSON: `{
				"metadata": {"name": "test-pod", "labels": {"environment": "preprod"}},
				"spec": {"containers":[{"name":"nginx","image":"nginx:latest"}],"restartPolicy":"Never"}
			}`,
			expectedPatch: patchJSON,
			expectAllowed: true,
		},
		{
			name: "do not patch annotated pod",
			podJSON: `{
				"metadata": {"name":"test-pod","labels":{"environment":"test"},"annotations":{"mutation.webhook.admission.k8s.io/skip":"true"}},
				"spec":{"containers":[{"name":"nginx","image":"nginx:latest"}],"restartPolicy":"Never"}
			}`,
			expectedPatch: "",
			expectAllowed: true,
		},
		{
			name: "do not patch pod with Job owner",
			podJSON: `{
				"metadata":{"name":"test-pod","labels":{"environment":"test"},"ownerReferences":[{"apiVersion":"batch/v1","kind":"Job","name":"example-job","uid":"job-123"}]},
				"spec":{"containers":[{"name":"nginx","image":"nginx:latest"}],"restartPolicy":"Never"}
			}`,
			expectedPatch: "",
			expectAllowed: true,
		},
		{
			name: "do not patch pod with CronJob owner",
			podJSON: `{
				"metadata":{"name":"test-pod","labels":{"environment":"test"},"ownerReferences":[{"apiVersion":"batch/v1","kind":"CronJob","name":"example-cron-job","uid":"job-123"}]},
				"spec":{"containers":[{"name":"nginx","image":"nginx:latest"}],"restartPolicy":"Never"}
			}`,
			expectedPatch: "",
			expectAllowed: true,
		},
		{
			name: "do not patch pod with dev env",
			podJSON: `{
				"metadata":{"name":"test-pod","labels":{"environment":"dev"}},
				"spec":{"containers":[{"name":"nginx","image":"nginx:latest"}],"restartPolicy":"Never"}
			}`,
			expectedPatch: "",
			expectAllowed: true,
		},
		{
			name: "do not patch pod missing env label",
			podJSON: `{
				"metadata":{"name":"test-pod"},
				"spec":{"containers":[{"name":"nginx","image":"nginx:latest"}],"restartPolicy":"Never"}
			}`,
			expectedPatch: "",
			expectAllowed: true,
		},
		{
			name: "do not patch pod with correct policy",
			podJSON: `{
				"metadata":{"name":"test-pod","labels":{"environment":"preprod"}},
				"spec":{"containers":[{"name":"nginx","image":"nginx:latest"}],"restartPolicy":"Always"}
			}`,
			expectedPatch: "",
			expectAllowed: true,
		},
		{
			name: "patch pod with no policy",
			podJSON: `{
				"metadata":{"name":"test-pod","labels":{"environment":"preprod"}},
				"spec":{"containers":[{"name":"nginx","image":"nginx:latest"}]}
			}`,
			expectedPatch: patchJSON,
			expectAllowed: true,
		},
		{
			name: "patch pod with no metadata",
			podJSON: `{
				"spec": {"containers": [{"name": "nginx","image":"nginx:latest"}],"restartPolicy":"Never"}
			}`,
			expectedPatch: "",
			expectAllowed: true,
		},
		{
			name: "patch pod with no containers",
			podJSON: `{
				"metadata":{"name":"test-pod","labels":{"environment":"test"}},
				"spec":{"restartPolicy":"Never"}
			}`,
			expectedPatch: patchJSON,
			expectAllowed: true,
		},
		{
			name: "patch pod with unexpected owner kind",
			podJSON: `{
				"metadata":{
					"name":"test-pod",
					"labels":{"environment":"test"},
					"ownerReferences":[{"apiVersion":"custom/v1","kind":"SomethingElse","name":"example","uid":"abc"}]
				},
				"spec":{"containers":[{"name":"nginx","image":"nginx:latest"}],"restartPolicy":"Never"}
			}`,
			expectedPatch: patchJSON,
			expectAllowed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runMutatePodTest(t, tt.podJSON, tt.expectedPatch, tt.expectAllowed)
		})
	}
}
