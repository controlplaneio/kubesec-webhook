package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	kubesecv2 "github.com/controlplaneio/kubectl-kubesec/v2/pkg/kubesec"
	"github.com/slok/kubewebhook/v2/pkg/log"
	"github.com/slok/kubewebhook/v2/pkg/model"
	"github.com/slok/kubewebhook/v2/pkg/webhook"
	"github.com/slok/kubewebhook/v2/pkg/webhook/validating"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kjson "k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/client-go/kubernetes/scheme"
)

// kubesecValidator validates a definition against the Kubesec.io score.
type validator struct {
	minScore int
	logger   log.Logger
}

var _ validating.Validator = &validator{}

// New returns a new validating webhook.
func New(minScore int, logger log.Logger) (webhook.Webhook, error) {
	val := &validator{
		minScore: minScore,
		logger:   logger,
	}

	cfg := validating.WebhookConfig{
		ID:        "kubesec",
		Validator: val,
		Logger:    logger,
	}

	return validating.NewWebhook(cfg)
}

// Validate implements the validator interface to validate a resource.
func (v *validator) Validate(_ context.Context, _ *model.AdmissionReview, obj metav1.Object) (*validating.ValidatorResult, error) {
	// Make sure the input object is a supported resource
	var kObj runtime.Object
	switch o := obj.(type) {
	case *appsv1.DaemonSet:
		kObj = o
	case *appsv1.Deployment:
		kObj = o
	case *v1.Pod:
		kObj = o
	case *appsv1.StatefulSet:
		kObj = o
	default:
		return &validating.ValidatorResult{
			Valid:    true,
			Warnings: []string{"resource kind not supported, validation skipped"},
		}, nil
	}

	// Logging
	logFields := map[string]interface{}{
		"kind":      kObj.GetObjectKind().GroupVersionKind().Kind,
		"namespace": obj.GetNamespace(),
		"name":      obj.GetName(),
	}
	logger := v.logger.WithValues(logFields)

	// Serialize the runtime object to yaml
	serializer := kjson.NewYAMLSerializer(kjson.DefaultMetaFactory, scheme.Scheme, scheme.Scheme)
	buf := bytes.NewBuffer([]byte{})
	err := serializer.Encode(kObj, buf)
	if err != nil {
		return &validating.ValidatorResult{}, fmt.Errorf("serialization failed: %w", err)
	}

	// Scan using kubesec.io
	logger.WithValues(log.Kv{"status": "running"}).Infof("scan resource")
	results, err := scan(*buf)
	if err != nil {
		logger.WithValues(log.Kv{"status": "failed", "error": err}).Infof("scan finished with error")
		return &validating.ValidatorResult{}, err
	}
	result := results[0]

	jq, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return &validating.ValidatorResult{}, fmt.Errorf("kubesec.io pretty printing issue %v", err)
	}

	logger.WithValues(log.Kv{"status": "success", "result": string(jq)}).Infof("scan finished successfuly")

	if result.Score < v.minScore {
		return &validating.ValidatorResult{
			Valid:   false,
			Message: fmt.Sprintf("%s score is %d, minimum accepted score is %d\nScan Result:\n%s", obj.GetName(), result.Score, v.minScore, jq),
		}, nil
	}

	return &validating.ValidatorResult{Valid: true}, nil
}

// scan is a small wrapper for scanning a manifest definition against kubesec.io
func scan(definition bytes.Buffer) ([]kubesecv2.KubesecResult, error) {
	results, err := kubesecv2.NewClient(kubesecScanURL, timeOut).
		ScanDefinition(definition)
	if err != nil {
		return results, fmt.Errorf("kubesec.io scan failed: %w", err)
	}

	if len(results) != 1 {
		return results, fmt.Errorf("scan failed as result is empty")
	}

	if results[0].Error != "" {
		return results, fmt.Errorf("kubesec.io scan failed: %s", results[0].Error)
	}

	return results, nil
}
