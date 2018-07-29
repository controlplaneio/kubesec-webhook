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
	"github.com/stefanprodan/kubectl-kubesec/pkg/kubesec"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kjson "k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/client-go/kubernetes/scheme"
)

// podValidator validates the definition against the Kubesec.io score.
type podValidator struct {
	minScore int
	logger   log.Logger
}

func (d *podValidator) Validate(_ context.Context, obj metav1.Object) (bool, validating.ValidatorResult, error) {
	kObj, ok := obj.(*v1.Pod)
	if !ok {
		return false, validating.ValidatorResult{Valid: true}, nil
	}

	serializer := kjson.NewYAMLSerializer(kjson.DefaultMetaFactory, scheme.Scheme, scheme.Scheme)
	var buffer bytes.Buffer
	writer := bufio.NewWriter(&buffer)

	kObj.TypeMeta = metav1.TypeMeta{
		Kind:       "Pod",
		APIVersion: "v1",
	}

	err := serializer.Encode(kObj, writer)
	if err != nil {
		d.logger.Errorf("pod serialization failed %v", err)
		return false, validating.ValidatorResult{Valid: true}, nil
	}

	writer.Flush()

	d.logger.Infof("Scanning pod %s", kObj.Name)

	result, err := kubesec.NewClient().ScanDefinition(buffer)
	if err != nil {
		d.logger.Errorf("kubesec.io scan failed %v", err)
		return false, validating.ValidatorResult{Valid: true}, nil
	}
	if result.Error != "" {
		d.logger.Errorf("kubesec.io scan failed %v", result.Error)
		return false, validating.ValidatorResult{Valid: true}, nil
	}

	if result.Score < d.minScore {
		return true, validating.ValidatorResult{
			Valid:   false,
			Message: fmt.Sprintf("%s score is %d, minimum accepted score is %d", kObj.Name, result.Score, d.minScore),
		}, nil
	}

	return false, validating.ValidatorResult{Valid: true}, nil
}

// NewPodWebhook returns a new deployment validating webhook.
func NewPodWebhook(minScore int, mrec metrics.Recorder, logger log.Logger) (webhook.Webhook, error) {

	// Create validators.
	val := &podValidator{
		minScore: minScore,
		logger:   logger,
	}

	cfg := validating.WebhookConfig{
		Name: "kubesec-pod",
		Obj:  &v1.Pod{},
	}

	return validating.NewWebhook(cfg, val, mrec, logger)
}
