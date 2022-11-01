package k3s

import (
	"fmt"
	"gopkg.in/yaml.v3"
)

const (
	DefaultConfigPath = "/etc/rancher/k3s/config.yaml"
)

type Config struct {
	WriteKubeConfigMode string   `yaml:"write-kubeconfig-mode"`
	TLSSan              []string `yaml:"tls-san"`
}

func NewConfig() *Config {
	return &Config{WriteKubeConfigMode: "0775"}
}

func (c *Config) SetWriteKubeConfigMode(mode string) *Config {
	c.WriteKubeConfigMode = mode
	return c
}

func (c *Config) SetTLSSan(nameOrAddr ...string) *Config {
	c.TLSSan = append(c.TLSSan, nameOrAddr...)
	return c
}

func (c *Config) ToYAML() (string, error) {
	d, err := yaml.Marshal(c)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s", d), nil
}
