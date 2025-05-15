package main

import "net/http"

func main() {

	server := http.Server{
		Addr:    "localhost:8080",
		Handler: nil,
	}
	server.ListenAndServe()
	// http.ListenAndServe("localhost:8080", nil)

}
