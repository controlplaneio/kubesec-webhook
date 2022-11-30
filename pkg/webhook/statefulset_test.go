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

// Test_statefulsetValidator_Validate - tests the validation of hardened and insecure YAML manifests
// The hardened manifest should be allowed by the webhook and the insecure should be blocked
func Test_statefulsetValidator_Validate(t *testing.T) {
	tests := []struct {
		name            string // name of the test
		wantErr         bool   // are we expecting an error
		result          bool   // response/result we expect from the webhook
		minScore        int    // minimum score used for initialisation
		statefulsetSpec string // Statefulset specification in string
	}{
		{
			name:     "Hardened Statefulset Spec",
			wantErr:  false,
			result:   true,
			minScore: 0,
			statefulsetSpec: `
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: hardened-statefulset
spec:
  selector:
    matchLabels:
      app: hardened-statefulset
  serviceName: "statefulset-test-sa"
  replicas: 2
  template:
    metadata:
      labels:
        app: hardened-statefulset
    spec:
      containers:
      - name: main-container
        image: nginx
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
			name:     "Insecure Statefulset Spec",
			wantErr:  false,
			result:   false,
			minScore: 0,
			statefulsetSpec: `
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: web
spec:
  selector:
    matchLabels:
      app: insecure-statefulset
  serviceName: "statefulset-test-sa"
  replicas: 2
  template:
    metadata:
      labels:
        app: insecure-statefulset
    spec:
      containers:
      - name: main-container
        image: nginx
        securityContext:
          readOnlyRootFilesystem: false
          runAsUser: 0
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

			pv := statefulSetValidator{
				minScore: tt.minScore,
				logger:   log.Dummy,
			}

			decoder := serializer.NewCodecFactory(scheme.Scheme).UniversalDecoder()

			deploy := &appsv1.StatefulSet{}

			if err := runtime.DecodeInto(decoder, []byte(tt.statefulsetSpec), deploy); err != nil {
				t.Fatalf("unable to convert %q into StatefulSet object - %v", tt.statefulsetSpec, err)
			}

			_, resp, err := pv.Validate(context.Background(), deploy)

			if (err != nil) != tt.wantErr {
				t.Fatalf("Statefulset validator - got error %v, but wanted %v", err, tt.wantErr)
			}

			got := resp.Valid
			want := tt.result

			if got != want {
				t.Fatalf("Statefulset validator - result mismatch, want=%v, got=%v", want, got)
				return
			}

		})
	}
}
