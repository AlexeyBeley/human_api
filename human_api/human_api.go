package human_api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/AlexeyBeley/human_api/azure_devops_api"
)

type Configuration struct {
	SprintName             string `json:"sprint_name"`
	ReportsDirPath         string `json:"reports_dir_path"`
	WorkerId               string `json:"worker_id"`
	AzureDevopsKey         string `json:"azure_devops_key"`
	AzureDevopsCompanyName string `json:"azure_devops_company"`
}

const preReportFileName = "pre_report.json"
const inputFileName = "input.hapi"
const baseFileName = "base.hapi"
const postReportFileName = "post_report.json"

func DailyRoutine(configFilePath string) error {
	/*
		if _, err:= os.Stat(reportFilePath) ; err == nil {
			fmt.Println("File exists")
		} else if os.IsNotExist(err) {
			fmt.Println("File does not exist")
		} else {
			fmt.Println("Error checking file existence:", err)
		}

	*/

	config, err := loadConfiguration(configFilePath)
	if err != nil {
		log.Printf("Failed with error: %v\n", err)
		return err
	}

	now := time.Now()
	dateDirName := now.Format("02-21")

	dateDirPath := filepath.Join(config.ReportsDirPath, config.SprintName, dateDirName)
	err = os.MkdirAll(filepath.Dir(dateDirPath), 0755)
	if err != nil {
		return err
	}

	preReportFilePath := filepath.Join(dateDirPath, preReportFileName)
	inputFilePath := filepath.Join(dateDirPath, inputFileName)
	baseFilePath := filepath.Join(dateDirPath, baseFileName)
	postReportFilePath := filepath.Join(dateDirPath, postReportFileName)

	if _, err := os.Stat(postReportFilePath); err == nil {
		return fmt.Errorf("post report file exists. The routine finished: %v", dateDirPath)
	}

	azure_devops_config := azure_devops_api.Configuration{PersonalAccessToken: config.AzureDevopsKey, OrganizationName: config.AzureDevopsCompanyName}
	if !checkFileExists(inputFilePath) {
		return DailyRoutineExtract(azure_devops_config, preReportFilePath, inputFilePath, baseFilePath, postReportFilePath)
	}

	return DailyRoutineSubmit(azure_devops_config, preReportFilePath, inputFilePath, baseFilePath, postReportFilePath)

}

func DailyRoutineExtract(config azure_devops_api.Configuration, preReportFilePath, inputFilePath, baseFilePath, postReportFilePath string) (err error) {
	if !checkFileExists(preReportFilePath) {
		if checkFileExists(inputFilePath) {
			return fmt.Errorf("pre report file does not exist. Input file exists '%v'", inputFilePath)
		}
		if checkFileExists(baseFilePath) {
			return fmt.Errorf("pre report file does not exist. Base file exists '%v'", baseFilePath)
		}
		DownloadSprintStatus(config, preReportFilePath)
	}

	if !checkFileExists(inputFilePath) {
		ConvertDailyJsonToHR(preReportFilePath, baseFilePath)
		err = copyFile(baseFilePath, inputFilePath)
		if err != nil {
			fmt.Println("Error copying file:", err)
			return err
		}
		return nil
	} else if checkFileExists(baseFilePath) {
		return fmt.Errorf("input file does not exist. Base file exists '%v'", baseFilePath)
	}

	if _, err := os.Stat(preReportFilePath); err == nil {
		fmt.Println("File exists")
	} else if os.IsNotExist(err) {
		fmt.Println("File does not exist")

		//ConvertToHapi(filepath.Dir(reportFilePath))
	} else {
		fmt.Println("Error checking file existence:", err)
	}
	return nil
}

func DailyRoutineSubmit(config azure_devops_api.Configuration, preReportFilePath, inputFilePath, baseFilePath, postReportFilePath string) (err error) {
	if !checkFileExists(preReportFilePath) ||
		!checkFileExists(inputFilePath) ||
		!checkFileExists(baseFilePath) ||
		checkFileExists(postReportFilePath) {
		return fmt.Errorf("undefined status: %s", postReportFilePath)
	}

	inputJsonFilePath := strings.Replace(filepath.Base(inputFilePath), ".hapi", "_hapi.json", 1)

	reports, err := ConvertHRToDailyJson(inputFilePath, inputJsonFilePath)
	if err != nil{
		return err
	}
	logWithLineNumber(fmt.Sprintf("Submitted %d", len(reports)))

	azure_devops_api.SubmitSprintStatus(config, []azure_devops_api.Wobject{})
	return fmt.Errorf("todo: implement")
}

func loadConfiguration(filePath string) (config Configuration, err error) {
	config = Configuration{SprintName: "",
		ReportsDirPath:         "",
		WorkerId:               "",
		AzureDevopsKey:         "",
		AzureDevopsCompanyName: ""}

	filePath, err = filepath.Abs(filePath)
	if err != nil {
		return config, err
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return config, err
	}

	err = json.Unmarshal(data, &config)
	if err != nil {
		return config, err
	}

	return config, nil

}

func DownloadSprintStatus(config azure_devops_api.Configuration, dstFilePath string) (err error) {
	log.Printf("%v, %v", config, dstFilePath)
	err = azure_devops_api.DownloadSprintStatus(config)
	return err
}

//Return True if exists, False if not or fails on error.
func checkFileExists(path string) (exists bool) {
	_, err := os.Stat(path)
	if err != nil {
		return true
	} else if os.IsNotExist(err) {
		return false
	}

	log.Fatalf("Failed checking file exists: %v", err)
	return false
}

func copyFile(srcFilePath, dstFilePath string) error {
	// Open the source file for reading
	srcFile, err := os.Open(srcFilePath)
	if err != nil {
		fmt.Println("Error opening source file:", err)
		return err
	}
	defer srcFile.Close()

	// Create the destination file (with 0644 permissions)
	dstFile, err := os.Create(dstFilePath)
	if err != nil {
		fmt.Println("Error creating destination file:", err)
		return err
	}
	defer dstFile.Close()

	// Copy the contents from source to destination
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		fmt.Println("Error copying file:", err)
		return err
	}

	return nil
}

func logWithLineNumber(message string) {
	// Get the caller's file name and line number
	_, file, line, ok := runtime.Caller(1) // 1 skips the current function
	if !ok {
		file = "???"
		line = 0
	}

	// Format the log message with line number
	logMessage := fmt.Sprintf("%s:%d: %s", file, line, message)

	// Print the log message
	fmt.Println(logMessage)
}
