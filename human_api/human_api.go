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
	SprintName                       string `json:"SprintName"`
	ReportsDirPath                   string `json:"ReportsDirPath"`
	WorkerId                         string `json:"WorkerId"`
	AzureDevopsConfigurationFilePath string `json:"AzureDevopsConfigurationFilePath"`
}

type Wobject struct {
	Id           string   `json:"Id"`
	Title        string   `json:"Title"`
	Description  string   `json:"Description"`
	LeftTime     int      `json:"LeftTime"`
	InvestedTime int      `json:"InvestedTime"`
	WorkerID     string   `json:"WorkerID"`
	ChildrenIDs  []string `json:"ChildrenIDs"`
	ParentID     string   `json:"ParentID"`
}

const preReportFileName = "pre_report.json"
const inputFileName = "input.hapi"
const baseFileName = "base.hapi"
const postReportFileName = "post_report.json"

func check(e error) {
	if e != nil {
		panic(e)
	}
}

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
	fmt.Println("starting daily routine")
	config, err := loadConfiguration(configFilePath)
	if err != nil {
		log.Printf("Failed with error: %v\n", err)
		return err
	}
	fmt.Println("Loaded config")

	now := time.Now()
	dateDirName := now.Format("2006_01_02")

	dateDirPath := filepath.Join(config.ReportsDirPath, config.SprintName, dateDirName)
	fmt.Println("Generated new directory path: " + dateDirPath)

	curDir, err := os.Getwd()
	check(err)
	fmt.Printf("%v", curDir)

	os.Chdir(filepath.Join(config.ReportsDirPath, config.SprintName))

	//err = os.MkdirAll(filepath.Dir(dateDirPath), 0755)
	err = os.MkdirAll(dateDirName, 0755)
	if err != nil {
		fmt.Printf("was not able to create '%v'", dateDirPath)
		return err
	}
	os.Chdir(curDir)

	fmt.Println("Created new directory path: " + dateDirPath)

	preReportFilePath := filepath.Join(dateDirPath, preReportFileName)
	inputFilePath := filepath.Join(dateDirPath, inputFileName)
	baseFilePath := filepath.Join(dateDirPath, baseFileName)
	postReportFilePath := filepath.Join(dateDirPath, postReportFileName)

	if _, err := os.Stat(postReportFilePath); err == nil {
		return fmt.Errorf("post report file exists. The routine finished: %v", dateDirPath)
	}

	azure_devops_config, err := azure_devops_api.LoadConfig(config.AzureDevopsConfigurationFilePath)
	if err != nil {
		return err
	}
	log.Printf("inputFilePath: %v", inputFilePath)
	if !checkFileExists(inputFilePath) {
		return DailyRoutineExtract(config, azure_devops_config, preReportFilePath, inputFilePath, baseFilePath, postReportFilePath)
	}

	return DailyRoutineSubmit(azure_devops_config, preReportFilePath, inputFilePath, baseFilePath, postReportFilePath)

}

func DailyRoutineExtract(config Configuration, azureDevopsConfig azure_devops_api.Configuration, preReportFilePath, inputFilePath, baseFilePath, postReportFilePath string) (err error) {
	if !checkFileExists(preReportFilePath) {
		if checkFileExists(inputFilePath) {
			return fmt.Errorf("pre report file does not exist. Input file exists '%v'", inputFilePath)
		}
		if checkFileExists(baseFilePath) {
			return fmt.Errorf("pre report file does not exist. Base file exists '%v'", baseFilePath)
		}
		DownloadAllWits(azureDevopsConfig, preReportFilePath)
	}

	if !checkFileExists(inputFilePath) {

		dailyJSONFilePath, err := GenerateDailyRepprt(config, preReportFilePath)
		check(err)
		_, err = ConvertDailyJsonToHR(dailyJSONFilePath, baseFilePath)
		check(err)

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

func GenerateDailyRepprt(config Configuration, statusFilePath string) (dstFilePath string, err error) {
	wobjects, err := ConvertAzureDevopsStatusToWobjects(statusFilePath)
	check(err)
	reportFilePath, err := GenerateDailyReport(config, wobjects)
	log.Printf("%v", reportFilePath)
	return reportFilePath, err
	//WorkerDailyReport{}
}

func GenerateDailyReport(config Configuration, wobjects []Wobject) (reportFilePath string, err error) {
	log.Printf("%v", wobjects)
	return reportFilePath, nil
}

func ConvertAzureDevopsStatusToWobjects(filePath string) (wobjects []Wobject, err error) {
	log.Printf("%v", filePath)
	return wobjects, nil
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
	if err != nil {
		return err
	}
	logWithLineNumber(fmt.Sprintf("Submitted %d", len(reports)))

	azure_devops_api.SubmitSprintStatus(config, []azure_devops_api.Wobject{})
	return fmt.Errorf("todo: implement")
}

func loadConfiguration(filePath string) (config Configuration, err error) {
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

func DownloadAllWits(config azure_devops_api.Configuration, dstFilePath string) (err error) {
	log.Printf("%v, %v", config, dstFilePath)
	err = azure_devops_api.DownloadAllWits(config, dstFilePath)
	return err
}

// Return True if exists, False if not or fails on error.
func checkFileExists(path string) (exists bool) {
	_, err := os.Stat(path)
	if err == nil {
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
