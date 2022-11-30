package webhook

import (
	"context"
	"testing"

	"github.com/slok/kubewebhook/pkg/log"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
)

// Test_deploymentValidator_Validate - tests the validation of hardened and insecure Deployment YAML manifests
// The hardened manifest should be allowed by the webhook and the insecure should be blocked
func Test_deploymentValidator_Validate(t *testing.T) {
	tests := []struct {
		name           string // name of the test
		wantErr        bool   // are we expecting an error
		result         bool   // response/result we expect from the webhook
		minScore       int    // minimum score used for initialisation
		deploymentSpec string // deployment specification in string
	}{
		{
			name:     "Hardened Deployment Spec",
			wantErr:  false,
			result:   true, // should be allowed by the webhook
			minScore: 0,
			deploymentSpec: `
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: hardened-deployment
spec:
  selector:
    matchLabels:
      app: foo
  replicas: 1
  template:
    metadata:
      labels:
        app: foo
    spec:
      containers:
      - name: main-container
        image: quay.io/fluentd_elasticsearch/fluentd:v2.5.2
        securityContext:
          readOnlyRootFilesystem: true
          runAsUser: 100
          runAsNonRoot: true
          privileged: false
          allowPrivilegeEscalation: false
          capabilities:
            drop:
            - "ALL"
        resources:
          limits:
            memory: 200Mi
          requests:
            cpu: 100m
            memory: 200Mi
`,
		},
		{
			name:     "Insecure Deployment Spec",
			wantErr:  false,
			result:   false, // should be blocked by the webhook
			minScore: 0,
			deploymentSpec: `
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-test
spec:
  selector:
    matchLabels:
      app: nginx
  replicas: 1
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: main-container
        image: quay.io/fluentd_elasticsearch/fluentd:v2.5.2
        securityContext:
          readOnlyRootFilesystem: false
          runAsUser: 100
          runAsNonRoot: false
          privileged: true
          allowPrivilegeEscalation: true

`,
		},
	}
	for _, tt := range tests {

		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pv := deploymentValidator{
				minScore: tt.minScore,
				logger:   log.Dummy,
			}

			decoder := serializer.NewCodecFactory(scheme.Scheme).UniversalDecoder()

			deploy := &appsv1.Deployment{}

			if err := runtime.DecodeInto(decoder, []byte(tt.deploymentSpec), deploy); err != nil {
				t.Fatalf("unable to convert %q into Deployment object - %v", tt.deploymentSpec, err)
			}

			_, resp, err := pv.Validate(context.Background(), deploy)

			if (err != nil) != tt.wantErr {
				t.Fatalf("Deployment validator - got error %v, but wanted %v", err, tt.wantErr)
			}

			got := resp.Valid
			want := tt.result

			if got != want {
				t.Fatalf("Deployment validator - result mismatch, want=%v, got=%v", want, got)
			}

		})
	}
}
