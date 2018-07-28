package main

import (
	"fmt"
	"log"
	"net/http"
)

var (
	//contain key=md5 and value=path of file
	hashmap  map[string]string
	datapath string
)

func main() {
	port := 8080
	http.HandleFunc("/read", readfileHandler)
	http.HandleFunc("/upload", uploadfileHandler)
	http.HandleFunc("/delete", deletefileHandler)
	hashmap = make(map[string]string)
	datapath = "data/"
	log.Printf("Server starting on port %v\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), nil))
}

//http://127.0.0.1/read?file=IWantToDownloadThisFile.zip
func readfileHandler(w http.ResponseWriter, r *http.Request) {
	readfile := NewDataFile()
	readfile.Read(w, r)
}

func uploadfileHandler(w http.ResponseWriter, r *http.Request) {
	uploadFile := NewDataFile()
	uploadFile.Write(w, r)
}

func deletefileHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello World\n")
}
