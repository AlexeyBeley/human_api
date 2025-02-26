package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/AlexeyBeley/human_api/human_api"
)

func main() {
	action := flag.String("action", "none", "The action to take")
	configFilePath := flag.String("cfg", "none", "Configuration file path")
	flag.Parse()

	if *action == "daily" {
		human_api.DailyRoutine(*configFilePath)
	} else if *action == "none" {
		log.Println("none")
	} else {
		log.Fatalf("Unknown action '%v'", *action)
	}

	fmt.Print("here")
}
