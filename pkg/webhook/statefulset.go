package webhook

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/slok/kubewebhook/pkg/log"
	"github.com/slok/kubewebhook/pkg/observability/metrics"
	"github.com/slok/kubewebhook/pkg/webhook"
	"github.com/slok/kubewebhook/pkg/webhook/validating"
	"github.com/controlplaneio/kubectl-kubesec/pkg/kubesec"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kjson "k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/client-go/kubernetes/scheme"
	"encoding/json"
)

// statefulSetValidator validates the definition against the Kubesec.io score.
type statefulSetValidator struct {
	minScore int
	logger   log.Logger
}

func (d *statefulSetValidator) Validate(_ context.Context, obj metav1.Object) (bool, validating.ValidatorResult, error) {
	kObj, ok := obj.(*appsv1beta1.StatefulSet)
	if !ok {
		d.logger.Errorf("received invalid StatefulSet object %v", obj)
		return false, validating.ValidatorResult{Valid: true}, nil
	}

	serializer := kjson.NewYAMLSerializer(kjson.DefaultMetaFactory, scheme.Scheme, scheme.Scheme)
	var buffer bytes.Buffer
	writer := bufio.NewWriter(&buffer)

	kObj.TypeMeta = metav1.TypeMeta{
		Kind:       "StatefulSet",
		APIVersion: "apps/v1",
	}

	err := serializer.Encode(kObj, writer)
	if err != nil {
		d.logger.Errorf("statefulset serialization failed %v", err)
		return false, validating.ValidatorResult{Valid: true}, nil
	}

	writer.Flush()

	d.logger.Infof("Scanning statefulset %s", kObj.Name)

	result, err := kubesec.NewClient().ScanDefinition(buffer)
	if err != nil {
		d.logger.Errorf("kubesec.io scan failed %v", err)
		return false, validating.ValidatorResult{Valid: true}, nil
	}
	if result.Error != "" {
		d.logger.Errorf("kubesec.io scan failed %v", result.Error)
		return false, validating.ValidatorResult{Valid: true}, nil
	}

	jq, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
	    d.logger.Errorf("kubesec.io pretty printing issue %v", err)
	    return false, validating.ValidatorResult{Valid: true}, nil
	}
	d.logger.Infof("Scan Result:\n%s", jq)

	if result.Score < d.minScore {
		return true, validating.ValidatorResult{
			Valid:   false,
			Message: fmt.Sprintf("%s score is %d, statefulset minimum accepted score is %d\nScan Result:\n%s", kObj.Name, result.Score, d.minScore, jq),
		}, nil
	}

	return false, validating.ValidatorResult{Valid: true}, nil
}

// NewStatefulSetWebhook returns a new statefulset validating webhook.
func NewStatefulSetWebhook(minScore int, mrec metrics.Recorder, logger log.Logger) (webhook.Webhook, error) {

	// Create validators.
	val := &statefulSetValidator{
		minScore: minScore,
		logger:   logger,
	}

	cfg := validating.WebhookConfig{
		Name: "kubesec-statefulset",
		Obj:  &appsv1beta1.StatefulSet{},
	}

	return validating.NewWebhook(cfg, val, mrec, logger)
}
