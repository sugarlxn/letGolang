package controller

import (
	"fmt"
	"net/http"
	"regexp"
)

func registerCommanyRoutes() {
	// Register all routes here
	http.HandleFunc("/companies", handlecommpanies)
	http.HandleFunc("/company/", handlecompany)
}

func handlecommpanies(w http.ResponseWriter, r *http.Request) {
	// Handle the request for companies
	fmt.Fprintf(w, "List of companies: google, apple, microsoft")
}

func handlecompany(w http.ResponseWriter, r *http.Request) {
	// Handle the request for a specific company
	pattern, _ := regexp.Compile(`/company/(\d+)`)
	matches := pattern.FindStringSubmatch(r.URL.Path)
	if len(matches) > 1 {
		companyID := matches[1]
		fmt.Fprintf(w, "Details of company with ID: %s", companyID)
	} else {
		http.NotFound(w, r)
	}
}
