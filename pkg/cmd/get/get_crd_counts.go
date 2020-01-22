package get

import (
	"fmt"
	"github.com/jenkins-x/jx/pkg/log"
	"sort"
	"strconv"

	"github.com/jenkins-x/jx/pkg/cmd/helper"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/jenkins-x/jx/pkg/cmd/opts"
	"github.com/jenkins-x/jx/pkg/cmd/templates"
)

// CRDCountOptions the command line options
type CRDCountOptions struct {
	*opts.CommonOptions
}

type tableLine struct {
	name    string
	version string
	count   int
}

var (
	getCrdCountLong = templates.LongDesc(`
		Count the number of resources for all custom resources definitions

`)

	getCrdCountExample = templates.Examples(`

		# Count the number of resources for all custom resources definitions
		jx get crd count
	`)
)

// NewCmdGetCount creates the command object
func NewCmdGetCRDCount(commonOpts *opts.CommonOptions) *cobra.Command {
	options := &CRDCountOptions{
		CommonOptions: commonOpts,
	}

	cmd := &cobra.Command{
		Use:     "crd count",
		Short:   "Display resources count for all custom resources",
		Long:    getCrdCountLong,
		Example: getCrdCountExample,
		Run: func(cmd *cobra.Command, args []string) {
			options.Cmd = cmd
			options.Args = args
			err := options.Run()
			helper.CheckErr(err)
		},
	}
	return cmd
}

// Run implements this command
func (o *CRDCountOptions) Run() error {
	results, err := o.getCustomResourceCounts()
	if err != nil {
		return errors.Wrap(err, "cannot get custom resource counts")
	}

	table := o.CreateTable()
	table.AddRow("NAME", "VERSION", "COUNT")

	for _, r := range results {
		table.AddRow(r.name, r.version, strconv.Itoa(r.count))
	}

	table.Render()
	return nil
}

func (o *CRDCountOptions) getCustomResourceCounts() ([]tableLine, error) {

	log.Logger().Info("this operation may take a while depending on how many custom resources exist")

	exClient, err := o.ApiExtensionsClient()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get api extensions client")
	}
	dynamicClient, _, err := o.GetFactory().CreateDynamicClient()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get dynamic client")
	}

	// lets loop over each arg and validate they are resources, note we could have "--all"
	crdList, err := exClient.ApiextensionsV1beta1().CustomResourceDefinitions().List(v1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get a list of custom resource definitions")
	}

	var results []tableLine
	// loop over each crd and check how many resources exist for them
	for _, crd := range crdList.Items {
		// each crd can have multiple versions
		for _, v := range crd.Spec.Versions {
			r := schema.GroupVersionResource{Group: crd.Spec.Group, Version: v.Name, Resource: crd.Spec.Names.Plural}

			resources, err := dynamicClient.Resource(r).List(v1.ListOptions{})
			if err != nil {
				return nil, errors.Wrapf(err, "finding resource %s.%s %s", crd.Spec.Names.Plural, crd.Spec.Group, v.Name)
			}

			line := tableLine{
				name:    fmt.Sprintf("%s.%s", crd.Spec.Names.Plural, crd.Spec.Group),
				version: v.Name,
				count:   len(resources.Items),
			}
			results = append(results, line)
		}
	}
	// sort the entries so resources with the most come at the bottom as it's clearer to see after running the command
	sort.Slice(results, func(i, j int) bool {
		return results[i].count < results[j].count
	})
	return results, nil
}
