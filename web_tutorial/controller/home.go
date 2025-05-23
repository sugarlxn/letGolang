package controller

import "net/http"

func registerhome() {
	http.HandleFunc("/home", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("home page"))
	})
}
