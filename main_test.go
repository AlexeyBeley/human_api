package main

import (
	"log"
	"os"
	"path/filepath"
	"testing"
)

var testConfig = Configuration{
	SprintName:             "sp1",
	ReportsDirPath:         "1",
	WorkerId:               "Horey1",
	AzureDevopsKey:         "secret",
	AzureDevopsCompanyName: "company",
}

func TestLoadConfiguration(t *testing.T) {
	t.Run("Init test", func(t *testing.T) {
		filePath, err := GetConfigFilePath("")
		if err != nil {
			t.Errorf("Failed to generate cofig file: %s", err)

		}
		config, err := loadConfiguration(filePath)
		if err != nil {
			t.Errorf("Failed with file: %s, %v", filePath, err)

		}
		if config.SprintName != testConfig.SprintName {
			t.Errorf("loadConfiguration() = %v, want %v", config.SprintName, testConfig.SprintName)
		}
	})
}

func GetConfigFilePath(basename string) (string, error) {
	cwd_path, err := os.Getwd()
	if err != nil {
		return "", err
	}
	abs_path, err := filepath.Abs(cwd_path)
	if err != nil {
		return "", err
	}
	if basename == "" {
		basename = "config.json"
	}
	dst_file_path := filepath.Join(abs_path, basename)
	log.Printf("Generated test destination HAPI file path: %s", dst_file_path)
	return dst_file_path, nil

}
