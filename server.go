package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

var (
	//contain key=md5 and value=path of file
	//hashmap  map[string]string
	datapath  string
	hashMongo HashDAO
)

func main() {

	r := mux.NewRouter()

	port := 8080
	r.HandleFunc("/read", readfileHandler)
	r.HandleFunc("/upload", uploadfileHandler)
	r.HandleFunc("/delete", deletefileHandler)

	//hashmap = make(map[string]string)
	hashMongo := HashDAO{"mongodb://mongoadmin:secret@127.0.0.1/?connect=direct&authMechanism=SCRAM-SHA-1", "storage"}
	hashMongo.Connect()

	datapath = "data/"
	log.Printf("Server starting on port %v\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), r))
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
	uploadFile := NewDataFile()
	uploadFile.Delete(w, r)
}
