package main

import (
	"fmt"

	"github.com/AlexeyBeley/human_api/human_api"
)

func main() {
	human_api.DailyRoutine()
	/*
		   daily_handler --action daily_json_to_hr --src --dst

		action := flag.String("action", "none", "The action to take")

		flag.Parse()

		if *action == "daily" {
			log.Println("daily")
		} else if *action == "hr_to_daily_json" {
			log.Println("daily")
		} else {
			log.Fatalf("Unknown action '%v'", *action)
		}
	*/
	fmt.Print("here")
}
