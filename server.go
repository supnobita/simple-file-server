package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	port := 8080
	http.HandleFunc("/read", readfileHandler)
	http.HandleFunc("/upload", uploadfileHandler)
	http.HandleFunc("/delete", deletefileHandler)
	log.Printf("Server starting on port %v\n", 8080)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), nil))
}
func readfileHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello World\n")
}

func uploadfileHandler(w http.ResponseWriter, r *http.Request) {

	uploadFile := DataFile{}
	uploadFile.Write(w, r)
}

func deletefileHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello World\n")
}
