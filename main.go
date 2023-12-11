package main

import (
	"cmd/main.go/repositories"
	"fmt"
	"html/template"
	"log"
	"net/http"
)

func init() {
	repositories.LoadEnvVariables()
	repositories.ConnectToDatabase()
}

type HeaderLinks struct {
	Login   string
	Contact string
	About   string
}

func main() {
	fmt.Println("Starting TripFrame Server...")

	http.HandleFunc("/", homePage)
	http.HandleFunc("/login", loginPage)
	http.HandleFunc("/contact", contactPage)
	http.HandleFunc("/about", aboutPage)

	log.Fatal(http.ListenAndServe(":8000", nil))
}

func homePage(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	headerLinks := headerLinks()
	tmpl.Execute(w, headerLinks)
}

func loginPage(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/login/index.html"))
	headerLinks := headerLinks()
	tmpl.Execute(w, headerLinks)
}

// TODO - make changes to about index page
func aboutPage(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/about/index.html"))
	headerLinks := headerLinks()
	tmpl.Execute(w, headerLinks)
}

// TODO - make changes to contact index page
func contactPage(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/contact/index.html"))
	headerLinks := headerLinks()
	tmpl.Execute(w, headerLinks)
}

func headerLinks() map[string][]HeaderLinks {
	return map[string][]HeaderLinks{
		"HeaderLinksTab": {
			{Login: "Login", Contact: "Contact", About: "About"},
		},
	}
}
