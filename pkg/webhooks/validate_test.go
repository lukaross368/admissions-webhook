package webhooks

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

// helper to run a validate pod test
func runValidatePodTest(t *testing.T, podJSON string, expectedStatus int, expectAllowed bool) {
	ar := newAdmissionReview([]byte(podJSON))
	resp := ValidatePods(ar)

	require.Equal(t, expectedStatus, int(resp.Result.Code))
	require.Equal(t, expectAllowed, resp.Allowed)
}

func TestValidatePods(t *testing.T) {
	tests := []struct {
		name           string
		podJSON        string
		expectedStatus int
		expectAllowed  bool
	}{
		{
			name: "pod with no probes passes",
			podJSON: `{
				"metadata": {"name": "test-pod", "labels": {"environment": "test"}},
				"spec": {"containers":[{"name":"nginx","image":"nginx:latest"}]}
			}`,
			expectedStatus: http.StatusForbidden,
			expectAllowed:  false,
		},
		{
			name: "pod with liveness and readiness probes passes",
			podJSON: `{
				"metadata": {"name": "test-pod", "labels": {"environment": "test"}},
				"spec": {"containers":[{"name":"nginx","image":"nginx:latest",
					"livenessProbe":{"httpGet":{"path":"/","port":80}},
					"readinessProbe":{"httpGet":{"path":"/","port":80}}
				}]}
			}`,
			expectedStatus: http.StatusOK,
			expectAllowed:  true,
		},
		{
			name: "pod missing liveness probe fails",
			podJSON: `{
				"metadata": {"name": "test-pod"},
				"spec": {"containers":[{"name":"nginx","image":"nginx:latest",
					"readinessProbe":{"httpGet":{"path":"/","port":80}}
				}]}
			}`,
			expectedStatus: http.StatusForbidden,
			expectAllowed:  false,
		},
		{
			name: "pod missing readiness probe fails",
			podJSON: `{
				"metadata": {"name": "test-pod"},
				"spec": {"containers":[{"name":"nginx","image":"nginx:latest",
					"livenessProbe":{"httpGet":{"path":"/","port":80}}
				}]}
			}`,
			expectedStatus: http.StatusForbidden,
			expectAllowed:  false,
		},
		{
			name: "pod with no containers passes",
			podJSON: `{
				"metadata": {"name": "empty-pod"},
				"spec": {}
			}`,
			expectedStatus: http.StatusOK,
			expectAllowed:  true,
		},
		{
			name: "multi-container pod partial probes fail",
			podJSON: `{
				"metadata": {
					"name": "multi-container-pod",
					"labels": {"environment": "test"}
				},
				"spec": {
					"containers": [
						{"name": "nginx","image": "nginx:latest"},
						{"name": "sidecar-logger","image": "busybox","livenessProbe": {"exec": {"command": ["ls"]}}},
						{"name": "metrics","image": "prom/prometheus","livenessProbe": {"httpGet": {"path": "/metrics", "port": 9090}},"readinessProbe": {"httpGet": {"path": "/metrics", "port": 9090}}}
					]
				}
			}`,
			expectedStatus: http.StatusForbidden,
			expectAllowed:  false,
		},
		{
			name: "pod with multiple containers all probes",
			podJSON: `{
				"metadata":{"name":"multi-container-pod","labels":{"environment":"test"}},
				"spec":{
					"containers":[
						{"name":"nginx","image":"nginx:latest","livenessProbe":{"httpGet":{"path":"/","port":80}},"readinessProbe":{"httpGet":{"path":"/","port":80}}},
						{"name":"busybox","image":"busybox","livenessProbe":{"exec":{"command":["echo","ok"]}},"readinessProbe":{"exec":{"command":["echo","ok"]}}}
					]
				}
			}`,
			expectedStatus: http.StatusOK,
			expectAllowed:  true,
		},
		{
			name: "pod with multiple containers some missing probes",
			podJSON: `{
				"metadata":{"name":"multi-container-pod","labels":{"environment":"test"}},
				"spec":{
					"containers":[
						{"name":"nginx","image":"nginx:latest","livenessProbe":{"httpGet":{"path":"/","port":80}}},
						{"name":"busybox","image":"busybox","readinessProbe":{"exec":{"command":["echo","ok"]}}}
					]
				}
			}`,
			expectedStatus: http.StatusForbidden,
			expectAllowed:  false,
		},
		{
			name: "pod no metadata no containers",
			podJSON: `{
				"spec": {}
			}`,
			expectedStatus: http.StatusOK,
			expectAllowed:  true,
		},
		{
			name: "pod with unexpected owner reference",
			podJSON: `{
				"metadata":{
					"name":"pod-unexpected-owner",
					"ownerReferences":[{"apiVersion":"custom/v1","kind":"SomethingElse","name":"example","uid":"abc"}]
				},
				"spec":{
					"containers":[{"name":"nginx","image":"nginx:latest","livenessProbe":{"httpGet":{"path":"/","port":80}},"readinessProbe":{"httpGet":{"path":"/","port":80}}}]
				}
			}`,
			expectedStatus: http.StatusOK,
			expectAllowed:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runValidatePodTest(t, tt.podJSON, tt.expectedStatus, tt.expectAllowed)
		})
	}
}
