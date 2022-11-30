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

// Test_daemonValidator_Validate - tests the validation of hardened and insecure daemonset YAML manifests
// The hardened manifest should be allowed by the webhook and the insecure should be blocked
func Test_daemonValidator_Validate(t *testing.T) {
	tests := []struct {
		name     string // name of the test
		wantErr  bool   // are we expecting an error
		result   bool   // response/result we expect from the webhook
		minScore int    // minimum score used for initialisation
		dsSpec   string // DaemonSet specification in string
	}{
		{
			name:     "Hardened DaemonSet Spec",
			wantErr:  false,
			result:   true,
			minScore: 0,
			dsSpec: `
---
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: fluentd-elasticsearch
  namespace: kube-system
  labels:
    k8s-app: fluentd-logging
spec:
  selector:
    matchLabels:
      name: fluentd-elasticsearch
  template:
    metadata:
      labels:
        name: fluentd-elasticsearch
    spec:
      containers:
      - name: fluentd-elasticsearch
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
        volumeMounts:
        - name: varlog
          mountPath: /var/log
      terminationGracePeriodSeconds: 30
      volumes:
      - name: varlog
        hostPath:
          path: /var/log
`,
		},
		{
			name:     "Insecure DaemonSet Spec",
			wantErr:  false,
			result:   false,
			minScore: 0,
			dsSpec: `
---
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: fluentd-elasticsearch
  namespace: kube-system
  labels:
    k8s-app: fluentd-logging
spec:
  selector:
    matchLabels:
      name: fluentd-elasticsearch
  template:
    metadata:
      labels:
        name: fluentd-elasticsearch
    spec:
      containers:
      - name: fluentd-elasticsearch
        image: quay.io/fluentd_elasticsearch/fluentd:v2.5.2
        securityContext:
          readOnlyRootFilesystem: false
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
			pv := daemonSetsValidator{
				minScore: tt.minScore,
				logger:   log.Dummy,
			}

			decoder := serializer.NewCodecFactory(scheme.Scheme).UniversalDecoder()

			ds := &appsv1.DaemonSet{}

			if err := runtime.DecodeInto(decoder, []byte(tt.dsSpec), ds); err != nil {
				t.Fatalf("unable to convert %q into DaemonSet object - %v", tt.dsSpec, err)
			}

			_, resp, err := pv.Validate(context.Background(), ds)

			if (err != nil) != tt.wantErr {
				t.Fatalf("DaemonSet validator - got error %v, but wanted %v", err, tt.wantErr)
			}

			got := resp.Valid
			want := tt.result

			if got != want {
				t.Fatalf("DaemonSet validator - result mismatch, want=%v, got=%v", want, got)
			}

		})
	}
}
