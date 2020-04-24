package operator

import (
	"github.com/spf13/cobra"

	"github.com/damemi/opentelemetry-operator/pkg/operator"
	"github.com/damemi/opentelemetry-operator/pkg/version"
	"github.com/openshift/library-go/pkg/controller/controllercmd"
)

func NewOperator() *cobra.Command {
	cmd := controllercmd.
		NewControllerCommandConfig("opentelemetry-operator", version.Get(), operator.RunOperator).
		NewCommand()
	cmd.Use = "run"
	cmd.Short = "Start the OpenTelemetry operator"

	return cmd
}
