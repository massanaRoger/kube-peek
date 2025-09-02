package cmd

import (
	"github.com/massanaRoger/m/v2/internal/kube"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
)

type App struct {
	Client *kubernetes.Clientset
	root   *cobra.Command
}

func NewApp() (*App, error) {
	a := &App{}
	provider, err := kube.NewProvider()
	if err != nil {
		return nil, err
	}
	client, err := provider.ClientSet()
	if err != nil {
		return nil, err
	}

	a.Client = client

	a.root = &cobra.Command{
		Use:           "kubepeek",
		Short:         "kubepeek — list & query Kubernetes Pods (kubectl-lite)",
		Long:          "kubepeek is a kubectl-lite for listing & querying Pods with smart filters, wide output, JSON, and watch mode.",
		SilenceUsage:  true, // don't print usage on errors
		SilenceErrors: true, // let us format errors
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			return nil
		},
		// If run with no subcommand, show help.
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	getCmd := a.newGetCmd()
	getPodsCmd := a.newGetPodsCmd()

	getCmd.AddCommand(getPodsCmd)

	a.root.AddCommand(getCmd)

	return a, nil
}

func (a *App) newGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get",
		Short: "List a resource",
	}
}

func (a *App) newGetPodsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "pods",
		Short: "List pods",
		RunE: func(cmd *cobra.Command, args []string) error {
			return kube.ListPods(a.Client)
		},
	}
}

func Execute() error {
	app, err := NewApp()
	if err != nil {
		return err
	}
	err = app.root.Execute()
	if err != nil {
		return err
	}

	return nil
}
