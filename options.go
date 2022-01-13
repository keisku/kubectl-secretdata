package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/resource"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"sigs.k8s.io/yaml"
)

func NewCmd() *cobra.Command {
	o := NewOptions(genericclioptions.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	})

	cmd := &cobra.Command{
		Use:                   "secretdata [(-o|--output=json|yaml)] [NAME | -l label] ...) [flags]",
		DisableFlagsInUseLine: true,
		Short:                 "Display decoded secret data",
		Long: `Display decoded secret data.  Prints decoded secret data about the found
secrets. You can filter the list using a label selector and the --selector flag,
or using --regex. You will only see results in your current namespace unless
you pass --all-namespaces or --multi-namespaces.
`,
		Example: `
# List all secrets in json format
kubectl secretdata -A -o json

# List secrets in specified NAMESPACES in yaml form(default)
kubectl secretdata -m "ns1,ns2,ns3"

# List secrets which are matched with regex in specified NAMESPACE
kubectl secretdata -n ns1 --regex "^secret[0-9]"

# List secrets which are matched with regex in specified NAMESPACES
kubectl secretdata -multi-namespaces "ns1,ns2,ns3" --regex "^something"

# List secrets which are matched with labels from all namespaces
kubectl secretdata -A --selector "key1=value1,key2=value2"
`,
	}
	cmd.Flags().BoolVarP(&o.AllNamespaces, "all-namespaces", "A", o.AllNamespaces, "If present, list the requested object(s) across all namespaces. Namespace in current context is ignored even if specified with --namespace.")
	cmd.Flags().StringVarP(&o.MultiNamespacesString, "multi-namespaces", "m", o.MultiNamespacesString, `The multi namespacess separated by "," where secrets exist.`)
	cmd.Flags().StringVarP(&o.Output, "output", "o", o.Output, "The format of the result")
	cmd.Flags().StringVarP(&o.LabelSelector, "selector", "l", o.LabelSelector, "Selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2)")
	cmd.Flags().StringVar(&o.Regex, "regex", o.Regex, "The regular expression for secret name")
	kubeConfigFlags := genericclioptions.NewConfigFlags(true).WithDeprecatedPasswordFlag().WithDiscoveryBurst(300).WithDiscoveryQPS(50.0)
	flags := cmd.PersistentFlags()
	kubeConfigFlags.AddFlags(flags)
	matchVersionKubeConfigFlags := cmdutil.NewMatchVersionFlags(kubeConfigFlags)
	matchVersionKubeConfigFlags.AddFlags(flags)
	f := cmdutil.NewFactory(matchVersionKubeConfigFlags)

	cmd.Run = func(cmd *cobra.Command, args []string) {
		cmdutil.CheckErr(o.Complete(f, args))
		cmdutil.CheckErr(o.Run(f))
	}
	return cmd
}

type Options struct {
	genericclioptions.IOStreams

	// For flag parse
	Namespace             string
	AllNamespaces         bool
	MultiNamespacesString string
	Output                string
	LabelSelector         string
	Regex                 string

	// For Complete()
	SecretName      string
	CompliedRegex   *regexp.Regexp
	MultiNamespaces []string
}

func NewOptions(ios genericclioptions.IOStreams) *Options {
	return &Options{
		IOStreams: ios,
		// Default
		Output: "yaml",
		Regex:  ".*", // Match every words
	}
}

func (o *Options) Complete(f cmdutil.Factory, args []string) error {
	if 0 < len(args) {
		o.SecretName = args[0]
	}

	var err error
	var explicitNamespace bool
	o.Namespace, explicitNamespace, err = f.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return err
	}
	if !explicitNamespace {
		o.Namespace = ""
	}

	switch nTrue(o.Namespace != "", o.AllNamespaces, o.MultiNamespacesString != "") {
	case 0:
		return fmt.Errorf("must select at least a namespace")
	case 2, 3:
		return fmt.Errorf("must choose one option to use for selecting namespace")
	}
	if o.Namespace != "" {
		o.AllNamespaces = false
	}
	if o.MultiNamespacesString != "" {
		o.MultiNamespaces = strings.Split(o.MultiNamespacesString, ",")
	}
	if len(o.MultiNamespaces) == 1 {
		o.Namespace = o.MultiNamespaces[0]
		o.AllNamespaces = false
	}
	if 1 < len(o.MultiNamespaces) {
		o.AllNamespaces = true
	}

	if o.Output != "json" && o.Output != "yaml" {
		return fmt.Errorf(`%s is invalid: --output must be "yaml", or "json"`, o.Output)
	}

	o.CompliedRegex, err = regexp.Compile(o.Regex)
	if err != nil {
		return err
	}

	return nil
}

func nTrue(bools ...bool) int {
	n := 0
	for _, b := range bools {
		if b {
			n++
		}
	}
	return n
}

func (o *Options) Run(f cmdutil.Factory) error {
	infos, err := o.secretInfos(f)
	if err != nil {
		return err
	}

	secretdata := make(map[string]map[string]map[string]string, len(infos))
	for _, info := range infos {
		if !o.match(info.Namespace, info.Name) {
			continue
		}
		data, err := getSecretData(info.Object)
		if err != nil {
			return err
		}
		_, ok := secretdata[info.Namespace]
		if !ok {
			secretdata[info.Namespace] = make(map[string]map[string]string)
		}
		secretdata[info.Namespace][info.Name] = data
	}

	if len(secretdata) == 0 {
		return fmt.Errorf("No secrets found")
	}

	switch o.Output {
	case "json":
		v, err := json.MarshalIndent(secretdata, "", "    ")
		if err != nil {
			return err
		}
		fmt.Fprintf(o.Out, "%s\n", v)
	case "yaml":
		v, err := yaml.Marshal(secretdata)
		if err != nil {
			return err
		}
		fmt.Fprintf(o.Out, "%s\n", v)
	}

	return nil
}

// getSecretData get the decoded secret data.
func getSecretData(obj runtime.Object) (map[string]string, error) {
	switch t := obj.(type) {
	case *v1.Secret:
		d := make(map[string]string, len(t.Data))
		for k, v := range t.Data {
			d[k] = string(v)
		}
		return d, nil
	case *unstructured.Unstructured:
		u, ok := obj.(*unstructured.Unstructured)
		if !ok {
			return nil, fmt.Errorf("failed to convert runtime.Object to *unstructured.Unstructured: actual type is %s", reflect.TypeOf(obj))
		}
		m, ok := u.Object["data"].(map[string]interface{})
		if !ok {
			// This case means that Secret has no data.
			return nil, nil
		}
		data := make(map[string]string, len(m))
		for k, v := range m {
			b, err := base64.StdEncoding.DecodeString(v.(string))
			if err != nil {
				return nil, fmt.Errorf("decode %v: %w", v, err)
			}
			data[k] = string(b)
		}
		return data, nil
	}
	return nil, fmt.Errorf("%T is unexpected type", obj)
}

func (o *Options) secretInfos(f cmdutil.Factory) ([]*resource.Info, error) {
	builder := f.NewBuilder().
		Unstructured().
		NamespaceParam(o.Namespace).DefaultNamespace().
		AllNamespaces(o.AllNamespaces).
		LabelSelectorParam(o.LabelSelector)
	if o.SecretName == "" {
		builder.ResourceTypeOrNameArgs(true, "secret")
	} else {
		builder.ResourceTypeOrNameArgs(true, "secret", o.SecretName)
	}
	r := builder.ContinueOnError().
		Latest().
		Flatten().
		Do()
	if err := r.Err(); err != nil {
		return nil, err
	}

	return r.Infos()
}

func (o *Options) match(namespace, secretName string) bool {
	if !o.CompliedRegex.MatchString(secretName) {
		return false
	}
	if o.Namespace != "" {
		return o.Namespace == namespace
	}
	for _, ns := range o.MultiNamespaces {
		if ns == namespace {
			return true
		}
	}
	return o.AllNamespaces
}
