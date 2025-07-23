package main

type Kustomization struct {
	Resources    []string `yaml:"resources"`
	Transformers []string `yaml:"transformers,omitempty"`
}
