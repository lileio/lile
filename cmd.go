package lile

import "github.com/spf13/cobra"

// BaseCommand provides the basic flags vars for running a service
func BaseCommand(serviceName, shortDescription string) *cobra.Command {
	command := &cobra.Command{
		Use:   serviceName,
		Short: shortDescription,
	}

	command.PersistentFlags().StringVar(
		&service.Config.Host,
		"grpc-host",
		"0.0.0.0",
		"gRPC service hostname",
	)

	command.PersistentFlags().IntVar(
		&service.Config.Port,
		"grpc-port",
		8000,
		"gRPC port",
	)

	command.PersistentFlags().StringVar(
		&service.PrometheusConfig.Host,
		"prometheus-host",
		"0.0.0.0",
		"Prometheus metrics hostname",
	)

	command.PersistentFlags().IntVar(
		&service.PrometheusConfig.Port,
		"prometheus-port",
		9000,
		"Prometheus metrics port",
	)

	return command
}
