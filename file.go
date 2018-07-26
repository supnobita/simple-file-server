package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

type DataFile struct {
	duplication bool
	name        string
	path        string
	size        int64
}

func (df *DataFile) Read(w http.ResponseWriter) error {
	return nil
}

func (df *DataFile) Write(w http.ResponseWriter, r *http.Request) error {
	r.ParseMultipartForm(32 << 20)
	file, handler, err := r.FormFile("uploadfile")
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}
	defer file.Close()
	fmt.Fprintf(w, "%v", handler.Header)
	f, err := os.OpenFile("./data/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}
	defer f.Close()
	io.Copy(f, file)
	return nil
}

func (df *DataFile) Delete(w http.ResponseWriter, r *http.Request) error {
	return nil
}
