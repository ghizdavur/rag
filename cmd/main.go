package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"
)

type CronJobsListDetails struct {
	CronJobType string
	CronJobName string
}

func main() {
	fmt.Println("Starting Demand Cron Jobs Server...")

	// handler function #1 - returns the index.html template, with cronjob data
	h1 := func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("templates/index.html"))
		cronjobsListDetails := map[string][]CronJobsListDetails{
			"CronsList": {
				{CronJobType: "1", CronJobName: "15min CronJob"},
				{CronJobType: "2", CronJobName: "1h CronJob"},
				{CronJobType: "3", CronJobName: "4h CronJob"},
			},
		}
		tmpl.Execute(w, cronjobsListDetails)
	}

	// handler function #2 - returns the template block with the newly added cronjob, as an HTMX response
	h2 := func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1 * time.Second)
		cronjobType := r.PostFormValue("cronjobType")
		cronjobName := r.PostFormValue("cronjobName")
		tmpl := template.Must(template.ParseFiles("templates/index.html"))
		tmpl.ExecuteTemplate(w, "cron-jobs-list", CronJobsListDetails{CronJobType: cronjobType, CronJobName: cronjobName})
	}

	// define handlers
	http.HandleFunc("/", h1)
	http.HandleFunc("/add-cronjob/", h2)

	log.Fatal(http.ListenAndServe(":8000", nil))

}
