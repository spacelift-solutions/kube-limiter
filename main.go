package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
)

var kustomizationFilePath = "kustomization.yaml"
var allowedResourceTypesFilePath = "/mnt/workspace/.spacelift-kubernetes-allowed-resource-types"
var pruneWhiteListFilePath = "/mnt/workspace/.spacelift-kubernetes-prune-whitelist-resource-types"

func init() {
	setupLogger()
}

func main() {
	if !shouldContinue() {
		return
	}

	files, err := getFilesFromKustomization()
	if err != nil {
		logger.Fatalf("Error reading kustomization file: %v", err)
	}
	logger.Debugf("Files in kustomization.yaml: %v", files)

	allowedResourceTypes := newCommaSeparated()
	pruneWhiteList := newCommaSeparated()

	k := newKube()
	for _, file := range files {
		gvk, err := k.getGVK(file)
		if err != nil {
			logger.Fatalf("Error getting GVK for %s: %v", file, err)
		}

		resource, err := k.getResourceForGVK(gvk)
		if err != nil {
			logger.Fatalf("Error getting resource for GVK %s: %v", gvk, err)
		}
		logger.Debugf("Resource type for GVK %+v: %+v", gvk, resource)

		gv := resource.APIGroupVersion
		resourceType := fmt.Sprintf("%s.%s", resource.APIResource.Name, resource.APIResource.Group)
		if resource.APIGroup == "" {
			// If the group is empty, it means it's a core resource
			gv = fmt.Sprintf("core/%s", resource.APIGroupVersion)
			resourceType = resource.APIResource.Name
		}

		allowedResourceTypes.add(resourceType)
		pruneWhiteList.add(fmt.Sprintf("%s/%s", gv, resource.APIResource.Kind))
	}

	allowedResourceTypes.writeToFile(allowedResourceTypesFilePath)
	pruneWhiteList.writeToFile(pruneWhiteListFilePath)
}

func shouldContinue() bool {
	labels, ok := os.LookupEnv("TF_VAR_spacelift_stack_labels")
	if !ok {
		logger.Debugf("Environment variable TF_VAR_spacelift_stack_labels not set, skipping execution")
		return false
	}

	if strings.Contains(labels, "kube:destroy") {
		return false
	}

	return true
}

func getFilesFromKustomization() ([]string, error) {
	f, err := os.ReadFile(kustomizationFilePath)
	if err != nil {
		return nil, err
	}

	var config Kustomization
	err = yaml.Unmarshal(f, &config)
	if err != nil {
		logger.Fatalf("Error parsing YAML: %v", err)
	}

	return config.Resources, nil
}
