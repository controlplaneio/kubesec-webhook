package webhook

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/slok/kubewebhook/v2/pkg/log"
	kwhlogrus "github.com/slok/kubewebhook/v2/pkg/log/logrus"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

type testCase struct {
	name             string // name of the test
	valid            bool   // should the resource to be allowed
	minScore         int    // minimum score used for initialisation
	specFilepath     string // path to manifest file
	expectedWarnings int    // we expected warning for non supported resources
}

// TestValidate - tests the validation of hardened and insecure YAML manifests
// The hardened manifest should be allowed by the webhook and the insecure should be blocked
func TestValidate(t *testing.T) {
	tests := []testCase{
		{
			name:         "Hardened DaemonSet Spec",
			valid:        true,
			minScore:     0,
			specFilepath: "./testdata/daemonset-hardened.yaml",
		},
		{
			name:         "Insecure DaemonSet Spec",
			valid:        false,
			minScore:     0,
			specFilepath: "./testdata/daemonset-insecure.yaml",
		},
		{
			name:         "Hardened Deployment Spec",
			valid:        true,
			minScore:     0,
			specFilepath: "./testdata/deployment-hardened.yaml",
		},
		{
			name:         "Insecure Deployment Spec",
			valid:        false,
			minScore:     0,
			specFilepath: "./testdata/deployment-insecure.yaml",
		},
		{
			name:         "Hardened Pod Spec.",
			valid:        true,
			minScore:     0,
			specFilepath: "./testdata/pod-hardened.yaml",
		},
		{
			name:         "Insecure Pod Spec",
			valid:        false,
			minScore:     0,
			specFilepath: "./testdata/pod-insecure.yaml",
		},
		{
			name:         "Hardened Statefulset Spec",
			valid:        true,
			minScore:     0,
			specFilepath: "./testdata/statefulset-hardened.yaml",
		},
		{
			name:         "Insecure Statefulset Spec",
			valid:        false,
			minScore:     0,
			specFilepath: "./testdata/statefulset-insecure.yaml",
		},
		{
			name: "Unsupported resource",
			// we don't validate this kind of resource so we consider it valid
			valid:            true,
			minScore:         0,
			expectedWarnings: 1,
			specFilepath:     "./testdata/configmap.yaml",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			manifest, err := ioutil.ReadFile(filepath.Clean(tt.specFilepath))
			if err != nil {
				t.Fatalf("error opening fixture file: %v", err)
			}

			decode := scheme.Codecs.UniversalDeserializer().Decode
			obj, _, err := decode(manifest, nil, nil)

			if err != nil {
				t.Fatalf("Unable to decode YAML object: %s", err.Error())
			}

			switch o := obj.(type) {
			case *appsv1.DaemonSet:
				testValidator(t, o, tt)
			case *appsv1.Deployment:
				testValidator(t, o, tt)
			case *v1.Pod:
				testValidator(t, o, tt)
			case *appsv1.StatefulSet:
				testValidator(t, o, tt)
			case *v1.ConfigMap:
				testValidator(t, o, tt)
			default:
				t.Fatalf("resource kind not supported for testing")
			}
		})
	}
}

func testValidator(t *testing.T, obj metav1.Object, tt testCase) {
	var logger log.Logger
	if _, ok := os.LookupEnv("TEST_ENABLE_LOGGING"); ok {
		logrusLogEntry := logrus.NewEntry(logrus.New())
		logger = kwhlogrus.NewLogrus(logrusLogEntry)
	} else {
		loggerN, _ := test.NewNullLogger()
		logrusLogEntry := logrus.NewEntry(loggerN)
		logger = kwhlogrus.NewLogrus(logrusLogEntry)
	}

	v := validator{
		minScore: tt.minScore,
		logger:   logger,
	}

	resp, err := v.Validate(context.Background(), nil, obj)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err.Error())
	}

	if len(resp.Warnings) != tt.expectedWarnings {
		t.Errorf("Unexpected number of warnings, got=%d, want=%d", len(resp.Warnings), tt.expectedWarnings)
	}

	if resp.Valid != tt.valid {
		t.Fatalf("Invalid resource correctness, got=%v, want=%v", resp.Valid, tt.valid)
	}
}
