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
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kjson "k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/client-go/kubernetes/scheme"
)

// deploymentValidator will validate the replicas are between max and min (inclusive).
type deploymentValidator struct {
	minScore int
	logger   log.Logger
}

func (d *deploymentValidator) Validate(_ context.Context, obj metav1.Object) (bool, validating.ValidatorResult, error) {
	kObj, ok := obj.(*extensionsv1beta1.Deployment)
	if !ok {
		// If not a deployment just continue the validation chain(if there is one) and don't do nothing.
		return false, validating.ValidatorResult{Valid: true}, nil
	}

	serializer := kjson.NewYAMLSerializer(kjson.DefaultMetaFactory, scheme.Scheme, scheme.Scheme)
	var buffer bytes.Buffer
	writer := bufio.NewWriter(&buffer)

	kObj.TypeMeta = metav1.TypeMeta{
		Kind:       "Deployment",
		APIVersion: "apps/v1",
	}

	err := serializer.Encode(kObj, writer)
	if err != nil {
		d.logger.Errorf("deployment serialization failed %v", err)
		return false, validating.ValidatorResult{Valid: true}, nil
	}

	writer.Flush()

	d.logger.Infof("Scanning %s", kObj.Name)

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
			Message: fmt.Sprintf("%s score is %d, deployment minimum accepted score is %d", kObj.Name, result.Score, d.minScore),
		}, nil
	}

	return false, validating.ValidatorResult{Valid: true}, nil
}

// NewDeploymentWebhook returns a new deployment validating webhook.
func NewDeploymentWebhook(minScore int, mrec metrics.Recorder, logger log.Logger) (webhook.Webhook, error) {

	// Create validators.
	val := &deploymentValidator{
		minScore: minScore,
		logger:   logger,
	}

	cfg := validating.WebhookConfig{
		Name: "kubesec-deploymentValidator",
		Obj:  &extensionsv1beta1.Deployment{},
	}

	return validating.NewWebhook(cfg, val, mrec, logger)
}
