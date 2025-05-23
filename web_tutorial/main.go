package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"text/template"
	"web_tutorial/controller"
)

type myhander struct{}

func (h *myhander) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello world"))
}

type aboutHandler struct{}

func (h *aboutHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("about page"))
}

func upload_file(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.Write([]byte("Please use POST method to upload file"))
		return
	}

	err := r.ParseMultipartForm(32 << 20) // 32 MB limit
	if err != nil {
		w.Write([]byte("Error parsing form: " + err.Error()))
		return
	}

	files := r.MultipartForm.File["file"]
	if len(files) == 0 {
		w.Write([]byte("No file uploaded with name 'uploaded'"))
		return
	}

	fileHeader := files[0]
	file, err := fileHeader.Open()
	if err != nil {
		w.Write([]byte("Error opening file: " + err.Error()))
		return
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		w.Write([]byte("Error reading file: " + err.Error()))
		return
	}

	// 成功读取到文件内容，返回给客户端
	w.Write([]byte("File content:\n"))
	w.Write(data)
}

func tempalte_index(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("template.html")
	t.Execute(w, rand.Intn(20) > 10)
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

	http.HandleFunc("/upload", upload_file)

	http.HandleFunc("/template", tempalte_index)

	controller.RegisterRoutes()

	server.ListenAndServe()
	// http.ListenAndServe("localhost:8080", nil)
	// http.Handle()
	// http.DefaultServeMux
	// http.HandleFunc()
	// template.ParseFiles()

}
