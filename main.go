package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type Configuration struct {
	sprintName             string `json:"sprint_name"`
	reportsDirPath         string `json:"reports_dir_path"`
	workerId               string `json:"worker_id"`
	azureDevopsKey         string `json:"azure_devops_key"`
	azureDevopsCompanyName string `json:"azure_devops_company"`
}

const preReportFileName = "pre_report.json"
const inputFileName = "input.hapi"
const baseFileName = "base.hapi"
const postReportFileName = "post_report.json"

func main() {
	fmt.Println("Test")
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

	data, err := os.ReadFile(filePath)
	if err != nil {
		return Configuration{}, err
	}

	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return config, fmt.Errorf("error")

}
