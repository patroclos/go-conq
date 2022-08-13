package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"gopkg.in/yaml.v2"
)

var CurrentConfig = &AppConfigFile{Config: AppConfig{Profiles: map[string]struct{}{}}}

func init() {
	pth := filepath.Join(xdg.ConfigHome, "example.yaml")
	err := CurrentConfig.UnmarshalText([]byte(pth))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

type AppConfig struct {
	Profiles      map[string]struct{} `yaml:"profiles"`
	ActiveProfile string              `yaml:"activeProfile,omitempty"`
}

type AppConfigFile struct {
	Config AppConfig
}

func (ac *AppConfigFile) UnmarshalText(txt []byte) error {
	fil, err := os.ReadFile(string(txt))
	if err != nil {
		return fmt.Errorf("failed to read appconfig %q: %w", txt, err)
	}
	if err := yaml.Unmarshal(fil, &ac.Config); err != nil {
		return fmt.Errorf("failed reading yaml of appconfig-file %q: %w", txt, err)
	}
	return nil
}
