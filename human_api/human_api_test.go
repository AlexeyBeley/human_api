package human_api

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/AlexeyBeley/human_api/azure_devops_api"
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


func test_check(t *testing.T, err error) {
	if err != nil {
		t.Errorf("%v", err)
	}
}

func TestGenerateDailyReportFromWobjects(t *testing.T) {
	t.Run("Init test", func(t *testing.T) {
		filePath, err := GetConfigFilePath("")
		if err != nil {
			t.Errorf("Failed to generate cofig file: %s", err)

		}
		config, err := loadConfiguration(filePath)
		if err != nil {
			t.Errorf("Failed with file: %s, %v", filePath, err)

		}
		wobjects := map[string]*Wobject{"123":{
			Id:           "123",
			Title:        "Test Title",
			Description:  "Test Description",
			LeftTime:     1,
			InvestedTime: 2,
			WorkerID:     "Horey",
			ChildrenIDs:  &[]string{"1", "2"},
			ParentID:     "3",
		}}
		fileOutputPath := GenerateDailyReportFromWobjects(config, wobjects, "/tmp/base.hapi")
		if err != nil {
			t.Fatalf("%v", err)
		}
		log.Print(fileOutputPath)
	})
}

func TestConvertAzureDevopsStatusToWobjects(t *testing.T) {
	t.Run("Init test", func(t *testing.T) {

		wobjects, err := ConvertAzureDevopsStatusToWobjects("/tmp/wit.json")
		test_check(t, err)
		log.Printf("%v", wobjects)
	})
}

func TestGenerateDailyReport(t *testing.T) {
	t.Run("Init test", func(t *testing.T) {
		filePath, err := GetConfigFilePath("")
		test_check(t, err)
		config, err := loadConfiguration(filePath)
		test_check(t, err)
		GenerateDailyReport(config, "/tmp/wit.json", "/tmp/base.hapi")
		test_check(t, err)
	})
}


func TestDailyRoutine(t *testing.T) {
	t.Run("Init test", func(t *testing.T) {

		err := DailyRoutine("/tmp/human_api_config.json")
		if err != nil {
			t.Fatalf("%v", err)
		}
	})
}

func TestDailyRoutineSubmit(t *testing.T) {
	t.Run("Init test", func(t *testing.T) {
		filePath, err := GetConfigFilePath("")
		test_check(t, err)
		config, err := loadConfiguration(filePath)
		test_check(t, err)
		azure_devops_config, err := azure_devops_api.LoadConfig(config.AzureDevopsConfigurationFilePath)
		test_check(t, err)
		err = DailyRoutineSubmit(azure_devops_config, "/tmp/input.hapi", "/tmp/base.hapi", "/tmp/postSubmit.json")
		test_check(t, err)
	})
}
