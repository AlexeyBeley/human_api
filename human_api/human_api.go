package human_api

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
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

func DailyRoutine() {
	/*
		if _, err:= os.Stat(reportFilePath) ; err == nil {
			fmt.Println("File exists")
		} else if os.IsNotExist(err) {
			fmt.Println("File does not exist")
		} else {
			fmt.Println("Error checking file existence:", err)
		}
	*/
	azure_devops_api.DownloadSprintStatus()
	fmt.Printf("Test %s, %s, %s, %s", preReportFileName, inputFileName, baseFileName, postReportFileName)
	config, err := loadConfiguration("")
	if err != nil {
		log.Printf("%v", config)
	}

	now := time.Now()
	dateDirName := now.Format("02-21")

	reportFilePath := filepath.Join(config.ReportsDirPath, config.SprintName, dateDirName, preReportFileName)
	if _, err := os.Stat(reportFilePath); err == nil {
		fmt.Println("File exists")
	} else if os.IsNotExist(err) {
		fmt.Println("File does not exist")
		DownloadSprintStatus(config, reportFilePath)
		//ConvertToHapi(filepath.Dir(reportFilePath))
	} else {
		fmt.Println("Error checking file existence:", err)
	}

	/*
			if ! exists(after_report){
			validate_report()
			upload_report()
			download_sprint_status(after_report)
		} else {
			log.Fatalf("File exists: %s", after_report)
		}*/
}

func loadConfiguration(filePath string) (config Configuration, err error) {
	config = Configuration{SprintName: "",
		ReportsDirPath:         "",
		WorkerId:               "",
		AzureDevopsKey:         "",
		AzureDevopsCompanyName: ""}

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

func DownloadSprintStatus(config Configuration, dstFilePath string) (err error) {
	//azureDevopsAPI := AzureDevopsAPI{config.AzureDevopsCompanyName, config.AzureDevopsKey}
	//azureDevopsAPI.downloadSprintStatus(config.SprintName)
	log.Printf("%v, %v", config, dstFilePath)

	return err
}
