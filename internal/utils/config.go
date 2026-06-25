package utils

import (
	"gopkg.in/yaml.v3"
)

type FieldConfig struct {
	Name     string            `yaml:"name"`
	Label    map[string]string `yaml:"label"`
	Type     string            `yaml:"type"`
	Required bool              `yaml:"required"`
}

type StepConfig struct {
	Field string `yaml:"field"`
}

type CertificateTypeConfig struct {
	ID                  string            `yaml:"id"`
	DisplayName         map[string]string `yaml:"display_name"`
	Fields              []FieldConfig     `yaml:"fields"`
	Steps               []StepConfig      `yaml:"steps"`
	RequiresAttachments bool              `yaml:"requires_attachments"`
	MaxAttachments      int               `yaml:"max_attachments"`
}

type CertificateTypesConfig struct {
	Types []CertificateTypeConfig `yaml:"types"`
}

func ParseCertificateTypes(data []byte) (map[string]CertificateTypeConfig, error) {
	var cfg CertificateTypesConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	registry := make(map[string]CertificateTypeConfig)
	for _, t := range cfg.Types {
		registry[t.ID] = t
	}
	return registry, nil
}
