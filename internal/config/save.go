package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func Save(path string, cfg Config) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	defer file.Close()

	enc := yaml.NewEncoder(file)
	enc.SetIndent(2)
	defer enc.Close()
	if err := enc.Encode(cfg); err != nil {
		return fmt.Errorf("encode config: %w", err)
	}
	return nil
}

func AddUniqueInt(list []int, value int) []int {
	if value == 0 {
		return list
	}
	for _, existing := range list {
		if existing == value {
			return list
		}
	}
	return append(list, value)
}

func AddUniqueString(list []string, value string) []string {
	if value == "" {
		return list
	}
	for _, existing := range list {
		if existing == value {
			return list
		}
	}
	return append(list, value)
}
