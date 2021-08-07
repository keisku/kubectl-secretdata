package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	if err := newCmd().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func newCmd() *cobra.Command {
	o := newOptions()

	cmd := &cobra.Command{
		Use:          "kubectl-secret-data",
		Short:        "A better kubectl for finding decoded secret data",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			name := ""
			if len(args) == 1 {
				name = args[0]
			}
			return run(cmd.Context(), name, o)
		},
	}

	o.parseFlags(cmd)

	return cmd
}

type options struct {
	Namespace    string
	KubeContext  string
	KubeConfing  string
	Cluster      string
	Output       string
	Regex        string
	AllNamespace bool
}

func newOptions() options {
	return options{
		Namespace:    "",
		Output:       "yaml",
		AllNamespace: false,
	}
}

func (o *options) toKubectlOptions() []string {
	opts := []string{"-o", o.Output}
	if o.Cluster != "" {
		opts = append(opts, "--cluster", o.Cluster)
	}
	if o.KubeConfing != "" {
		opts = append(opts, "--kubeconfig", o.KubeConfing)
	}
	if o.KubeContext != "" {
		opts = append(opts, "--context", o.KubeContext)
	}
	return opts
}

func (o *options) parseFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(&o.Namespace, "namespace", "n", o.Namespace, `The namespaces where secrets exist. You can set multiple namespaces separated by ","`)
	cmd.PersistentFlags().StringVar(&o.KubeContext, "context", o.KubeContext, "The name of the kubeconfig cluster to use")
	cmd.PersistentFlags().StringVar(&o.Cluster, "cluster", o.Cluster, "The name of the kubeconfig context to use")
	cmd.PersistentFlags().StringVar(&o.KubeConfing, "kubeconfig", o.KubeConfing, "Path to the kubeconfig file to use for CLI requests")
	cmd.PersistentFlags().StringVarP(&o.Output, "output", "o", o.Output, "The format of the result")
	cmd.PersistentFlags().StringVarP(&o.Regex, "regex", "E", o.Regex, "The regular expression of secret name")
	cmd.PersistentFlags().BoolVarP(&o.AllNamespace, "all-namespace", "A", o.AllNamespace, "If present, find secrets from all namespaces")
}
