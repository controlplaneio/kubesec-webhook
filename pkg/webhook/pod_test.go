package webhook

import (
	"context"
	"testing"

	"github.com/slok/kubewebhook/pkg/log"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
)

// Test_daemonValidator_Validate - tests the validation of hardened and insecure Pod YAML manifests
// The hardened manifest should be allowed by the webhook and the insecure should be blocked
func Test_podValidator_Validate(t *testing.T) {
	tests := []struct {
		name     string // name of the test
		wantErr  bool   // are we expecting an error
		result   bool   // response/result we expect from the webhoo
		minScore int    // minimum score used for initialisation
		podSpec  string // pod specification in string
	}{
		{
			name:     "Hardened Pod Spec.",
			wantErr:  false,
			result:   true, // should be allowed by the webhook
			minScore: 0,
			podSpec: `
apiVersion: v1
kind: Pod
metadata:
  name: secure-pod-spec
  namespace: foo
spec:
  containers:
  - name: main
    image: busybox
    serviceAccount: test
    command: [ "sh", "-c", "sleep 1h" ]
    securityContext:
      readOnlyRootFilesystem: true
      runAsUser: 100
      runAsNonRoot: true
      privileged: false
      allowPrivilegeEscalation: false
      capabilities:
        drop:
          - "all"
`,
		},
		{
			name:     "Insecure Pod Spec",
			wantErr:  false,
			result:   false, // should be blocked by the webhook
			minScore: 0,
			podSpec: `
apiVersion: v1
kind: Pod
metadata:
  name: test
  namespace: foo
spec:
  containers:
  - name: main
    image: busybox
    serviceAccount: foo
    command: [ "sh", "-c", "sleep 1h" ]
    securityContext:
      readOnlyRootFilesystem: false
      privileged: true
      runAsNonRoot: false

`,
		},
	}
	for _, tt := range tests {

		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pv := podValidator{
				minScore: tt.minScore,
				logger:   log.Dummy,
			}

			decoder := serializer.NewCodecFactory(scheme.Scheme).UniversalDecoder()

			pod := &corev1.Pod{}

			if err := runtime.DecodeInto(decoder, []byte(tt.podSpec), pod); err != nil {
				t.Fatalf("unable to convert %q into Pod object - %v", tt.podSpec, err)
			}

			_, resp, err := pv.Validate(context.Background(), pod)

			if (err != nil) != tt.wantErr {
				t.Fatalf("Pod Validator - got error %v, but wanted %v", err, tt.wantErr)
			}

			got := resp.Valid
			want := tt.result

			if got != want {
				t.Fatalf("Pod Validator - result mismatch, want=%v, got=%v", want, got)
			}

		})
	}
}
