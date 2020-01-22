package get

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"testing"

	cmd_mocks "github.com/jenkins-x/jx/pkg/cmd/clients/mocks"
	. "github.com/petergtz/pegomock"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextentions_mocks "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	dymamic_mocks "k8s.io/client-go/dynamic/fake"
	kube_mocks "k8s.io/client-go/kubernetes/fake"

	"github.com/jenkins-x/jx/pkg/cmd/opts"
)

const (
	group = "wine.io"
)

func TestRun(t *testing.T) {
	o := CRDCountOptions{
		CommonOptions: &opts.CommonOptions{},
	}

	currentNamespace := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cellar",
			Namespace: "cellar",
		},
	}

	scheme := runtime.NewScheme()

	// setup mocks
	factory := cmd_mocks.NewMockFactory()
	kubernetesInterface := kube_mocks.NewSimpleClientset(currentNamespace)
	apiextensionsInterface := apiextentions_mocks.NewSimpleClientset(getClusterScopedCRD(), getNamespaceScopedCRD())
	dynamicInterface := dymamic_mocks.NewSimpleDynamicClient(scheme)
	r := schema.GroupVersionResource{Group: group, Version: "v1", Resource: "rioja"}

	_, err := dynamicInterface.Resource(r).Namespace("cellar").Create(getNamespaceResource("test1"), metav1.CreateOptions{})
	_, err = dynamicInterface.Resource(r).Namespace("cellar").Create(getNamespaceResource("test2"), metav1.CreateOptions{})

	r = schema.GroupVersionResource{Group: group, Version: "v1", Resource: "shiraz"}

	_, err = dynamicInterface.Resource(r).Create(getClusterResource("test3"), metav1.CreateOptions{})
	assert.NoError(t, err)

	// return our fake kubernetes client in the test
	When(factory.CreateKubeClient()).ThenReturn(kubernetesInterface, "cellar", nil)
	When(factory.CreateApiExtensionsClient()).ThenReturn(apiextensionsInterface, nil)
	When(factory.CreateDynamicClient()).ThenReturn(dynamicInterface, "cellar", nil)

	o.SetFactory(factory)

	// run the command
	rs, err := o.getCustomResourceCounts()
	assert.NoError(t, err)

	// the order is important here, larger counts should appear at the bottom of the list so we can see them sooner
	assert.Equal(t, 1, rs[0].count)
	assert.Equal(t, 2, rs[1].count)

}

func getNamespaceScopedCRD() *v1beta1.CustomResourceDefinition {
	return &v1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "rioja",
		},
		Spec: v1beta1.CustomResourceDefinitionSpec{
			Group: group,
			Versions: []v1beta1.CustomResourceDefinitionVersion{
				{
					Name: "v1",
				},
			},
			Scope: v1beta1.NamespaceScoped,
			Names: v1beta1.CustomResourceDefinitionNames{
				Plural: "rioja",
			},
		},
	}
}

func getClusterScopedCRD() *v1beta1.CustomResourceDefinition {
	return &v1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "shiraz",
		},
		Spec: v1beta1.CustomResourceDefinitionSpec{
			Group: group,
			Versions: []v1beta1.CustomResourceDefinitionVersion{
				{
					Name: "v1",
				},
			},
			Scope: v1beta1.ClusterScoped,
			Names: v1beta1.CustomResourceDefinitionNames{
				Plural: "shiraz",
			},
		},
	}
}

func getNamespaceResource(name string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "wine.io/v1",
			"kind":       "rioja",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": "cellar",
			},
		},
	}
}

func getClusterResource(name string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "wine.io/v1",
			"kind":       "shiraz",
			"metadata": map[string]interface{}{
				"name": name,
			},
		},
	}
}
