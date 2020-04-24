package operator

import (
	"context"
	"errors"
	"time"

	otelv1alpha1 "github.com/damemi/opentelemetry-operator/pkg/apis/otel/v1alpha1"
	operatorconfigclientv1beta1 "github.com/damemi/opentelemetry-operator/pkg/generated/clientset/versioned/typed/otel/v1alpha1"
	"github.com/damemi/opentelemetry-operator/pkg/operator/assets"
	"github.com/damemi/opentelemetry-operator/pkg/operator/operatorclient"

	operatorv1 "github.com/openshift/api/operator/v1"
	"github.com/openshift/library-go/pkg/controller/factory"
	"github.com/openshift/library-go/pkg/operator/events"
	"github.com/openshift/library-go/pkg/operator/resource/resourceapply"
	"github.com/openshift/library-go/pkg/operator/resource/resourcemerge"
	"github.com/openshift/library-go/pkg/operator/resource/resourceread"
	"github.com/openshift/library-go/pkg/operator/v1helpers"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/workqueue"
)

type TargetConfigReconciler struct {
	operatorClient operatorconfigclientv1beta1.OpentelemetriesV1alpha1Interface
	otelClient     *operatorclient.OpenTelemetryClient
	kubeClient     kubernetes.Interface
	eventRecorder  events.Recorder
	queue          workqueue.RateLimitingInterface
}

func NewTargetConfigReconciler(
	operatorConfigClient operatorconfigclientv1beta1.OpentelemetriesV1alpha1Interface,
	otelClient *operatorclient.OpenTelemetryClient,
	kubeClient kubernetes.Interface,
	eventRecorder events.Recorder,
) factory.Controller {
	c := &TargetConfigReconciler{
		operatorClient: operatorConfigClient,
		otelClient:     otelClient,
		kubeClient:     kubeClient,
	}

	return factory.New().WithInformers(
		otelClient.Informer(),
	).WithSync(c.sync).ResyncEvery(time.Second).ToController("TargetConfigController", eventRecorder.WithComponentSuffix("target-config-controller"))
}

func (c TargetConfigReconciler) sync(ctx context.Context, controllerContext factory.SyncContext) error {
	otel, err := c.operatorClient.OpenTelemetries(operatorclient.OperatorNamespace).Get(ctx, "collector", metav1.GetOptions{})
	if err != nil {
		return err
	}
	if len(otel.Spec.Config) == 0 {
		return errors.New("must have config set for OpenTelemetry Collector")
	}
	if len(otel.Spec.Service) == 0 {
		return errors.New("must configure at least one port for the OpenTelemetry Collector service")
	}

	_, _, err = c.manageService(ctx, controllerContext.Recorder(), otel)
	if err != nil {
		return err
	}

	forceDeployment := false
	cm, forceDeployment, err := c.manageConfigMap(ctx, controllerContext.Recorder(), otel)
	if err != nil {
		return err
	}

	deployment, _, err := c.manageDeployment(ctx, controllerContext.Recorder(), otel, forceDeployment, cm)
	if err != nil {
		return err
	}

	_, _, err = v1helpers.UpdateStatus(c.otelClient, func(status *operatorv1.OperatorStatus) error {
		resourcemerge.SetDeploymentGeneration(&status.Generations, deployment)
		return nil
	})
	return err
}

func (c *TargetConfigReconciler) manageConfigMap(ctx context.Context, recorder events.Recorder, otel *otelv1alpha1.OpenTelemetry) (*v1.ConfigMap, bool, error) {
	required := resourceread.ReadConfigMapV1OrDie(assets.MustAsset("bindata/configmap.yaml"))
	required.Data = map[string]string{"conf.yaml": otel.Spec.Config}
	return resourceapply.ApplyConfigMap(c.kubeClient.CoreV1(), recorder, required)
}

func (c *TargetConfigReconciler) manageService(ctx context.Context, recorder events.Recorder, otel *otelv1alpha1.OpenTelemetry) (*v1.Service, bool, error) {
	required := resourceread.ReadServiceV1OrDie(assets.MustAsset("bindata/service.yaml"))
	ports := make([]v1.ServicePort, 0, len(otel.Spec.Service))
	for _, svc := range otel.Spec.Service {
		ports = append(ports,
			v1.ServicePort{
				Name:       svc.Name,
				Port:       int32(svc.Port),
				TargetPort: intstr.FromInt(svc.TargetPort),
			})
	}
	required.Spec.Ports = ports
	return resourceapply.ApplyService(
		c.kubeClient.CoreV1(),
		recorder,
		required)
}

func (c *TargetConfigReconciler) manageDeployment(ctx context.Context, recorder events.Recorder, otel *otelv1alpha1.OpenTelemetry, forceDeployment bool, configMap *v1.ConfigMap) (*appsv1.Deployment, bool, error) {
	required := resourceread.ReadDeploymentV1OrDie(assets.MustAsset("bindata/deployment.yaml"))
	required.Spec.Template.Spec.Containers[0].Image = otel.Spec.Image

	return resourceapply.ApplyDeploymentWithForce(
		c.kubeClient.AppsV1(),
		recorder,
		required,
		resourcemerge.ExpectedDeploymentGeneration(required, otel.Status.Generations),
		forceDeployment)
}
