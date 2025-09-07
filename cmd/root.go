package cmd

import (
	"os"

	"github.com/massanaRoger/m/v2/internal/kube"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
)

type App struct {
	Client *kubernetes.Clientset
	root   *cobra.Command
	table  *tablewriter.Table
	flags  Flags
}

type Flags struct {
	namespace     string
	allNamespaces bool
	selector      string
	fieldSelector string
	watch         bool
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

	table := tablewriter.NewWriter(os.Stdout)

	a.Client = client

	a.table = table

	a.root = &cobra.Command{
		Use:           "kubepeek",
		Short:         "kubepeek — list & query Kubernetes Pods (kubectl-lite)",
		Long:          "kubepeek is a kubectl-lite for listing & querying Pods with smart filters, wide output, JSON, and watch mode.",
		SilenceUsage:  true, // don't print usage on errors
		SilenceErrors: true, // let us format errors
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			if a.flags.allNamespaces {
				a.flags.namespace = ""
			}
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

	a.root.PersistentFlags().StringVarP(&a.flags.namespace, "namespace", "n", "default", "The namespace scope for this CLI request")

	a.root.PersistentFlags().BoolVarP(&a.flags.allNamespaces, "all-namespaces", "A", false, "If present, list the requested object(s) across all namespaces. Namespace in current context is ignored even if specified with --namespace.")

	getCmd.PersistentFlags().StringVar(&a.flags.fieldSelector, "field-selector", "", "Selector (field query) to filter on, supports '=', '==', and '!='.(e.g. --field-selector key1=value1,key2=value2). The server only supports a limited number of field queries per type.")
	getCmd.PersistentFlags().StringVarP(&a.flags.selector, "selector", "l", "", "Selector (label query) to filter on, supports '=', '==', '!=', 'in', 'notin'.(e.g. -l key1=value1,key2=value2,key3 in (value3)). Matching objects must satisfy all of the specified label constraints")
	getCmd.PersistentFlags().BoolVarP(&a.flags.watch, "watch", "w", false, "After listing/getting the requested object, watch for changes")

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
			return kube.ListPods(a.Client, a.table, a.flags.namespace, kube.ListPodsFlags{Selector: a.flags.selector, FieldSelector: a.flags.fieldSelector, Watch: a.flags.watch})
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
