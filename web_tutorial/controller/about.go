package controller

import "net/http"

func registerabout() {
	http.HandleFunc("/about_com", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("about page"))
	})
}
