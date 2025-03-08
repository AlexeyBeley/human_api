package human_api

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"
)


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
		if config.SprintName != "" {
			t.Errorf("loadConfiguration() = %v", config.SprintName)
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
	fmt.Printf("Ignoring path: %v", dst_file_path)
	dst_file_path = "/tmp/human_api_config.json"
	log.Printf("Generated test destination HAPI file path: %s", dst_file_path)
	return dst_file_path, nil

}


func TestDailyRoutine(t *testing.T) {
	t.Run("Init test", func(t *testing.T) {

		err := DailyRoutine("/tmp/human_api_config.json")
		if err != nil{
			t.Fatalf("%v", err)
		}
	}) 
}