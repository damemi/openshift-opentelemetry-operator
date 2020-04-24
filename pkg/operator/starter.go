package operator

import (
	"context"
	"time"

	operatorconfigclient "github.com/damemi/opentelemetry-operator/pkg/generated/clientset/versioned"
	operatorclientinformers "github.com/damemi/opentelemetry-operator/pkg/generated/informers/externalversions"
	"github.com/damemi/opentelemetry-operator/pkg/operator/operatorclient"

	"github.com/openshift/library-go/pkg/controller/controllercmd"
	"github.com/openshift/library-go/pkg/operator/v1helpers"

	"k8s.io/client-go/kubernetes"
)

func RunOperator(ctx context.Context, cc *controllercmd.ControllerContext) error {
	kubeClient, err := kubernetes.NewForConfig(cc.ProtoKubeConfig)
	if err != nil {
		return err
	}

	kubeInformersForNamespaces := v1helpers.NewKubeInformersForNamespaces(kubeClient,
		"",
		operatorclient.OperatorNamespace,
	)

	operatorConfigClient, err := operatorconfigclient.NewForConfig(cc.KubeConfig)
	if err != nil {
		return err
	}
	operatorConfigInformers := operatorclientinformers.NewSharedInformerFactory(operatorConfigClient, 10*time.Minute)

	otelClient := &operatorclient.OpenTelemetryClient{
		SharedInformer: operatorConfigInformers.Opentelemetries().V1alpha1().OpenTelemetries().Informer(),
		OperatorClient: operatorConfigClient.OpentelemetriesV1alpha1(),
	}

	targetConfigReconciler := NewTargetConfigReconciler(
		operatorConfigClient.OpentelemetriesV1alpha1(),
		otelClient,
		kubeClient,
		cc.EventRecorder,
	)

	operatorConfigInformers.Start(ctx.Done())
	kubeInformersForNamespaces.Start(ctx.Done())
	targetConfigReconciler.Run(ctx, 1)

	return nil
}
