package main

import (
	"fmt"
	"net/http"
)

type myhander struct{}

func (h *myhander) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello world"))
}

type aboutHandler struct{}

func (h *aboutHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("about page"))
}

func main() {

	mh := myhander{}
	ah := aboutHandler{}
	server := http.Server{
		Addr:    "localhost:8080",
		Handler: nil,
	}
	http.Handle("/hello", &mh)
	http.Handle("/about", &ah)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("# root page"))
	},
	)

	//post test
	http.HandleFunc("/post", func(w http.ResponseWriter, r *http.Request) {
		length := r.ContentLength
		body := make([]byte, length)
		r.Body.Read(body)
		w.Write([]byte("post body: " + string(body)))
	})

	http.Handle("/aaa", http.NotFoundHandler())
	http.Handle("/bbb", http.RedirectHandler("http://www.bing.com", 302))

	http.HandleFunc("/submit", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		fmt.Fprintln(w, "name:", r.FormValue("name"))
		fmt.Fprintln(w, r.Form)
	})

	server.ListenAndServe()
	// http.ListenAndServe("localhost:8080", nil)
	// http.Handle()
	// http.DefaultServeMux
	// http.HandleFunc()

}
