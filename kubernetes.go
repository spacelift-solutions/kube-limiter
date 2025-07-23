package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
)

type kube struct {
	client       *kubernetes.Clientset
	apiResources []groupResource
}

type groupResource struct {
	APIGroup        string
	APIGroupVersion string
	APIResource     metav1.APIResource
}

func newKube() *kube {
	config, err := getKubeConfig()
	if err != nil {
		panic(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	lists, err := clientset.Discovery().ServerPreferredResources()
	if err != nil {
		logger.Fatalf("Error getting api resources: %v", err)
	}

	resources := []groupResource{}
	for _, list := range lists {
		if len(list.APIResources) == 0 {
			continue
		}
		gv, err := schema.ParseGroupVersion(list.GroupVersion)
		if err != nil {
			continue
		}
		for _, resource := range list.APIResources {
			resources = append(resources, groupResource{
				APIGroup:        gv.Group,
				APIGroupVersion: gv.String(),
				APIResource:     resource,
			})
		}
	}

	return &kube{
		client:       clientset,
		apiResources: resources,
	}
}

func (k kube) getGVK(fileName string) (schema.GroupVersionKind, error) {
	stream, err := os.ReadFile(fileName)
	if err != nil {
		return schema.GroupVersionKind{}, fmt.Errorf("error reading file %s: %v", fileName, err)
	}

	if len(stream) == 0 {
		return schema.GroupVersionKind{}, fmt.Errorf("file %s is empty", fileName)
	}

	type meta struct {
		Kind       string `yaml:"kind"`
		APIVersion string `yaml:"apiVersion"`
	}
	var metadata meta
	if err := yaml.Unmarshal(stream, &metadata); err != nil {
		return schema.GroupVersionKind{}, fmt.Errorf("error unmarshalling file %s: %v", fileName, err)
	}

	gvk := schema.FromAPIVersionAndKind(metadata.APIVersion, metadata.Kind)

	return gvk, nil
}

func (k kube) getResourceForGVK(gvk schema.GroupVersionKind) (groupResource, error) {
	for _, resource := range k.apiResources {
		if resource.APIResource.Kind == gvk.Kind && resource.APIGroupVersion == gvk.GroupVersion().String() {
			return resource, nil
		}
		if resource.APIResource.Kind == gvk.Kind && resource.APIGroup == "" && gvk.Group == "" {
			return resource, nil
		}
	}
	return groupResource{}, fmt.Errorf("resource type not found for GVK: %s", gvk)
}

func getKubeConfig() (*rest.Config, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	return kubeConfig.ClientConfig()
}
