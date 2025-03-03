/*
download_all go test 
*/
package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/AlexeyBeley/human_api/human_api"
	"github.com/AlexeyBeley/human_api/azure_devops_api"
)

func main() {
	action := flag.String("action", "none", "The action to take")
	configFilePath := flag.String("cfg", "none", "Configuration file path")
	flag.Parse()

	if *action == "daily" {
		human_api.DailyRoutine(*configFilePath)
	} else if *action == "download_all" {
		config, err := azure_devops_api.LoadConfig(*configFilePath)
		if err != nil{
			log.Fatalf("Error received '%v'", err)	
		}
		azure_devops_api.DownloadAllWits(config, "/tmp/wit.json")
	} else {
		log.Fatalf("Unknown action '%v'", *action)
	}

	fmt.Print("here")
}
