package operatorclient

import (
	"context"
	operatorconfigclientv1beta1 "github.com/damemi/opentelemetry-operator/pkg/generated/clientset/versioned/typed/otel/v1alpha1"
	operatorv1 "github.com/openshift/api/operator/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
)

const OperatorNamespace = "opentelemetry-operator"

type OpenTelemetryClient struct {
	SharedInformer cache.SharedIndexInformer
	OperatorClient operatorconfigclientv1beta1.OpentelemetriesV1alpha1Interface
}

func (c *OpenTelemetryClient) Informer() cache.SharedIndexInformer {
	return c.SharedInformer
}

func (c *OpenTelemetryClient) GetOperatorState() (spec *operatorv1.OperatorSpec, status *operatorv1.OperatorStatus, resourceVersion string, err error) {
	instance, err := c.OperatorClient.OpenTelemetries(OperatorNamespace).Get(context.TODO(), "collector", metav1.GetOptions{})
	if err != nil {
		return nil, nil, "", err
	}
	return &instance.Spec.OperatorSpec, &instance.Status.OperatorStatus, instance.ResourceVersion, nil
}

func (c *OpenTelemetryClient) UpdateOperatorSpec(resourceVersion string, spec *operatorv1.OperatorSpec) (out *operatorv1.OperatorSpec, newResourceVersion string, err error) {
	original, err := c.OperatorClient.OpenTelemetries(OperatorNamespace).Get(context.TODO(), "collector", metav1.GetOptions{})
	if err != nil {
		return nil, "", err
	}
	copy := original.DeepCopy()
	copy.ResourceVersion = resourceVersion
	copy.Spec.OperatorSpec = *spec

	ret, err := c.OperatorClient.OpenTelemetries(OperatorNamespace).Update(context.TODO(), copy, metav1.UpdateOptions{})
	if err != nil {
		return nil, "", err
	}

	return &ret.Spec.OperatorSpec, ret.ResourceVersion, nil
}

func (c *OpenTelemetryClient) UpdateOperatorStatus(resourceVersion string, status *operatorv1.OperatorStatus) (out *operatorv1.OperatorStatus, err error) {
	original, err := c.OperatorClient.OpenTelemetries(OperatorNamespace).Get(context.TODO(), "collector", metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	copy := original.DeepCopy()
	copy.ResourceVersion = resourceVersion
	copy.Status.OperatorStatus = *status

	ret, err := c.OperatorClient.OpenTelemetries(OperatorNamespace).UpdateStatus(context.TODO(), copy, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}

	return &ret.Status.OperatorStatus, nil
}
