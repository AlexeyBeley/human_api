package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
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

func main() {
	fmt.Printf("Test %s, %s, %s, %s", preReportFileName, inputFileName, baseFileName, postReportFileName)
	config, err := loadConfiguration("")
	if err != nil {
		log.Printf("%v", config)
	}
	/*
			if !exists (config.reports_dir_path / config.sprint_name / date_dir/ hapi_input ){
				download_sprint_status(before_report)
				convert_to_hapi()
				return
			}

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
