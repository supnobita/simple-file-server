package main

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
)

type DataFile struct {
	isduplicate bool
	name        string
	path        string
	size        int64
	md5hash     string
	refPath     string //path to original data, deduplication
	numOfRefer  int    //number of file that point to this file
}

func NewDataFile() DataFile {
	f := DataFile{}
	f.isduplicate = false
	f.name = ""
	f.path = ""
	f.size = -1
	f.md5hash = ""
	f.refPath = ""
	f.numOfRefer = 0
	return f
}

func (df *DataFile) Read(w http.ResponseWriter, r *http.Request) error {
	//First of check if Get is set in the URL
	Filename := r.URL.Query().Get("file")
	if Filename == "" {
		//Get not set, send a 400 bad request
		http.Error(w, "Get 'file' not specified in url.", http.StatusBadRequest)
		return fmt.Errorf("filename not found")
	}
	fmt.Println("Client requests: " + Filename)
	//load meta data and check if has reference file
	df.path = datapath + Filename
	//if can't load meta data, may be file not found
	if err := df.LoadMetaData(); err != nil {
		http.Error(w, "File not found", http.StatusBadRequest)
		return fmt.Errorf("Read metadata File not found: " + df.path)
	}
	if len(df.refPath) != 0 || df.isduplicate == true {
		//this file is duplicated file, we need to connect db find root path
		hash, err := hashMongo.FindHash(df.md5hash)
		if err != nil {
			log.Fatal("Error find hash from db: " + df.md5hash + " of file: " + df.path)
		}
		df.path = hash.Origpath
	}

	//Check if file exists and open
	Openfile, err := os.Open(df.path)
	defer Openfile.Close() //Close after function return
	if err != nil {
		//File not found, send 404
		http.Error(w, "File "+Filename+" not found.", http.StatusNotFound)
		fmt.Println("Read file error: " + df.path)
		return err
	}

	//File is found, create and send the correct headers

	//Get the Content-Type of the file
	//Create a buffer to store the header of the file in
	FileHeader := make([]byte, 512)
	//Copy the headers into the FileHeader buffer
	Openfile.Read(FileHeader)
	//Get content type of file
	FileContentType := http.DetectContentType(FileHeader)

	//Get the file size
	FileStat, _ := Openfile.Stat()                     //Get info from file
	FileSize := strconv.FormatInt(FileStat.Size(), 10) //Get file size as a string

	//Send the headers
	w.Header().Set("Content-Disposition", "attachment; filename="+Filename)
	w.Header().Set("Content-Type", FileContentType)
	w.Header().Set("Content-Length", FileSize)

	//Send the file
	//We read 512 bytes from the file already so we reset the offset back to 0
	Openfile.Seek(0, 0)
	io.Copy(w, Openfile) //'Copy' the file to the client
	return nil
}

func (df *DataFile) Write(w http.ResponseWriter, r *http.Request) error {
	r.ParseMultipartForm(32 << 20)
	file, handler, err := r.FormFile("uploadfile")
	if err != nil {
		fmt.Println("header error " + err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}
	defer file.Close()
	fmt.Fprintf(w, "%v", handler.Header)

	df.name = handler.Filename
	df.path = datapath + df.name

	//if file exist on server return
	if err := df.LoadMetaData(); err == nil {
		fmt.Println("file exist " + df.path)
		io.WriteString(w, "File "+df.name+" is Exist on Server, Use read api to get file")
		return err
	}

	f, err := os.OpenFile(df.path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("write file error: " + err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}
	defer f.Close()
	io.Copy(f, file)

	fileinfo, err := f.Stat()
	if err == nil {
		df.size = fileinfo.Size()
	}

	//gen hash of file
	df.CalculateMD5Hash()
	//md5value := ""
	//whether hash is valid or not
	if len(df.md5hash) > 0 {
		//check if this hash is exist
		//if we cannot find hash => good
		if hash, err := hashMongo.FindHash(df.md5hash); err == nil {
			if df.IsDuplicatedFile(hash.Origpath) {
				//load meta original file and add one refer
				df.isduplicate = true
				df.refPath = hash.Origpath
				originalMetaFile, err := LoadMetaData(df.refPath)
				if err == nil {
					originalMetaFile.numOfRefer++ // increase num of refer file
					//save to disk
					originalMetaFile.WriteMetaData()
					//write path whose file pointed to this originalMetaFile
					//originalMetaFile.WriteReferPathToMetaFile(df.path)
					DeleteFile(df.path)
				}
				//Append this path to hash.Referpaths array and Update hash object
				hash.Referpaths = append(hash.Referpaths, df.path)
				if err := hashMongo.UpdateHash(hash); err != nil {
					log.Fatal("Update hash error: " + err.Error())
					http.Error(w, "DB connection Error", http.StatusBadRequest)
					return err
				}
			} //if 2 file has different content, do nothing
		} else {
			//if hash is not exist, add it
			h := Hash{df.md5hash, df.path, []string{}}
			err := hashMongo.InsertHash(h)
			if err != nil {
				log.Fatal("Add hash error: " + err.Error())
				http.Error(w, "DB connection Error", http.StatusBadRequest)
				return err
			}
		}

	}
	if err := df.WriteMetaData(); err != nil {
		fmt.Println("write file error: " + err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}
	return nil
}

func (df *DataFile) IsDuplicatedFile(pathfile string) bool {
	//check file content of two file
	return deepCompare(pathfile, df.path)
}

func deepCompare(file1, file2 string) bool {
	// Check file size ...
	chunkSize := 4096
	f1, err := os.Open(file1)
	if err != nil {
		log.Fatal(err)
		return false
	}

	f2, err := os.Open(file2)
	if err != nil {
		log.Fatal(err)
		return false
	}
	defer f1.Close()
	defer f2.Close()
	for {
		b1 := make([]byte, chunkSize)
		_, err1 := f1.Read(b1)

		b2 := make([]byte, chunkSize)
		_, err2 := f2.Read(b2)

		if err1 != nil || err2 != nil {
			if err1 == io.EOF && err2 == io.EOF {
				return true
			} else if err1 == io.EOF || err2 == io.EOF {
				return false
			} else {
				log.Fatal(err1, err2)
			}
		}

		if !bytes.Equal(b1, b2) {
			return false
		}
	}
}

func DeleteFile(pathfile string) error {
	//rename file first then delete
	tempfile := ""
	//find the name of temp file
	for i := 0; i < 10; i++ {
		token := tokenGenerator()
		if IsFileExist(datapath+token+".temp") == false {
			tempfile = datapath + token + ".temp"
			break
		}
	}
	if len(tempfile) == 0 {
		tempfile = pathfile
	} else {
		//rename pathfile to tempfile
		if err := os.Rename(pathfile, tempfile); err != nil {
			fmt.Println("Rename file " + pathfile + " error: " + err.Error())
			//if rename error try to delete pathfile, if ok return nil
			if os.Remove(pathfile) == nil {
				return nil
			}
		}
	}
	//try to delete temp file
	return os.Remove(tempfile)
}

//IsFileExist method check file exist will return true
func (df *DataFile) IsFileExist() bool {
	return IsFileExist(df.path)
}

func IsFileExist(pathfile string) bool {
	if len(pathfile) == 0 {
		return false // return not exist
	}

	if _, err := os.Stat(pathfile); os.IsNotExist(err) {
		return false
	}
	return true //file exist
}

//CalculateMD5Hash will calculate md5 hash of file
func (df *DataFile) CalculateMD5Hash() {
	//if path dosen't contain filename or is empty, we should not calculate hash
	if len(df.path) < 4 {
		return
	}
	f, err := os.Open(df.path)
	if err != nil {
		fmt.Print("Calculate hash has error " + err.Error())
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		fmt.Print("Calculate hash has error " + err.Error())
	}

	df.md5hash = fmt.Sprintf("%x", h.Sum(nil))

}

func tokenGenerator() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

//WriteMetaData will write file meta data to disk ($filename.meta)
func (df *DataFile) WriteMetaData() error {
	data := df.name + "\n" + df.path + "\n" + strconv.FormatInt(df.size, 10) +
		"\n" + strconv.FormatBool(df.isduplicate) + "\n" + df.md5hash + "\n" + df.refPath + "\n" + strconv.Itoa(df.numOfRefer)

	f, err := os.OpenFile(df.path+".meta", os.O_WRONLY|os.O_CREATE, 0666)
	defer f.Close()
	if err != nil {
		return err
	}
	f.WriteString(data)
	return nil
}

func (df *DataFile) WriteReferPathToMetaFile(refpath string) error {
	// If the file doesn't exist, create it, or append to the file
	f, err := os.OpenFile(df.path+".meta", os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal("Write ReferPath to meta file " + df.path + " has error: " + err.Error())
	}
	if _, err := f.Write([]byte("\n" + refpath)); err != nil {
		log.Fatal("Write ReferPath to meta file " + df.path + " has error: " + err.Error())
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
	return err
}

//LoadMetaData load meta data file
func (df *DataFile) LoadMetaData() error {
	meta, err := LoadMetaData(df.path)
	if err == nil {
		df.name = meta.name
		df.size = meta.size
		df.isduplicate = meta.isduplicate
		df.md5hash = meta.md5hash
		df.numOfRefer = meta.numOfRefer
		df.refPath = meta.refPath
		//data := df.name + "--" + df.path + "--" + strconv.FormatInt(df.size, 10) +
		//	"--" + strconv.FormatBool(df.isduplicate) + "--" + df.md5hash + "--" + df.refPath + "--" + strconv.Itoa(df.numOfRefer)
		//fmt.Println("Loadmeta-test: " + data)
	}
	return err
}

//LoadMetaData load meta data file
func LoadMetaData(datafile string) (DataFile, error) {
	if len(datafile) == 0 {
		return NewDataFile(), os.ErrInvalid
	}
	//if file meta is not exist return error
	if IsFileExist(datafile+".meta") == false {
		return NewDataFile(), os.ErrNotExist
	}

	metadata := NewDataFile()
	metadata.path = datafile
	file, err := os.Open(metadata.path + ".meta")
	defer file.Close()
	if err != nil {
		log.Fatal(err)
		return metadata, err
	}
	scanner := bufio.NewScanner(file)
	data := make([]string, 7)
	for i := 0; scanner.Scan() && i < 7; i++ {
		data[i] = scanner.Text()
	}

	metadata.name = data[0]
	metadata.path = data[1]
	metadata.size, err = strconv.ParseInt(data[2], 10, 64)
	metadata.isduplicate, err = strconv.ParseBool(data[3])
	metadata.md5hash = data[4]
	metadata.refPath = data[5]
	metadata.numOfRefer, err = strconv.Atoi(data[6])
	return metadata, err
}

func (df *DataFile) Delete(w http.ResponseWriter, r *http.Request) error {
	//load meta of file
	Filename := r.URL.Query().Get("file")
	if Filename == "" {
		//Get not set, send a 400 bad request
		http.Error(w, "Get 'file' not specified in url.", http.StatusBadRequest)
		return fmt.Errorf("filename not found")
	}
	fmt.Println("Client requests: " + Filename)
	//load meta data and check if has reference file
	df.path = datapath + Filename
	if err := df.LoadMetaData(); err != nil {
		http.Error(w, "File not found", http.StatusBadRequest)
		return fmt.Errorf("Read metadata File not found: " + df.path)
	}

	//find hash key
	hash, err := hashMongo.FindHash(df.md5hash)
	if err != nil {
		http.Error(w, "DB connection Error", http.StatusInternalServerError)
		log.Fatal("Find hash " + df.md5hash + " error: " + err.Error())
		return err
	}

	//if it's orignial file
	if df.isduplicate == false {

		if len(hash.Referpaths) > 0 && hash.Origpath == df.path {
			//if hash is original ? and has more refer file
			// 1) chose next root file from it's refer files
			metafile := NewDataFile()
			for i := 0; i < len(hash.Referpaths); i++ {
				metafile, err := LoadMetaData(hash.Referpaths[i])
				if err != nil || metafile.path == "" { // if cannot load file, or meta is error
					continue
				} else {
					if metafile.path == "" {
						break
					}
					hash.Origpath = hash.Referpaths[i]
					//delete refer_path
					newReferPaths := append(hash.Referpaths[:i], hash.Referpaths[i+1:]...)
					hash.Referpaths = newReferPaths
					if err := hashMongo.UpdateHash(hash); err != nil {
						http.Error(w, "DB connection Error", http.StatusInternalServerError)
						log.Fatal("Update hash " + hash.Key + " error: " + err.Error())
						return err
					}
					// update meta data file of this refer_file
					metafile.isduplicate = false // change to root state
					metafile.refPath = ""
					metafile.numOfRefer = len(hash.Referpaths)
					//delete meta data of new file first then save new meta file
					DeleteFile(metafile.path + ".meta")
					metafile.WriteMetaData()
					//final we delete meta data of old root file and rename old root file to new name
					DeleteFile(df.path + ".meta")
					if err := os.Rename(df.path, metafile.path); err != nil {
						http.Error(w, "Server Error", http.StatusInternalServerError)
						log.Fatal("Rename file error " + metafile.path + " error: " + err.Error())
						return err
					}
					fmt.Println("file delte OK " + df.path)
					io.WriteString(w, "File "+df.name+" is deleted")
					return nil
				}
			}
			// if cannot find valid meta file or healthy refer file, treate as alone original file, delete
			if metafile.path == "" {
				if err := hashMongo.DeletetHashByKey(df.md5hash); err != nil {
					http.Error(w, "Delete File error", http.StatusBadRequest)
					return fmt.Errorf("Delete hash error: " + err.Error())
				}
				// final we delete meta file and file data
				DeleteFile(df.path + ".meta")
				DeleteFile(df.path)
				fmt.Println("file delte OK " + df.path)
				io.WriteString(w, "File "+df.name+" is deleted")
				return nil
			}

		} else { // if has numofrefer = 0 delete file

			// if this refer_paths == 0, it mean only one file map with this hash, delete
			if len(hash.Referpaths) == 0 && hash.Origpath == df.path {
				//delete hash key in mongo db
				if err := hashMongo.DeletetHashByKey(df.md5hash); err != nil {
					http.Error(w, "Delete File error", http.StatusBadRequest)
					return fmt.Errorf("Delete hash error: " + err.Error())
				}

			}
			// hash key is has more than 2 refer, may be this file is has same hash but content is difference
			// we will delete file only
			//delete meta file
			DeleteFile(df.path + ".meta")
			DeleteFile(df.path)
			fmt.Println("file delte OK " + df.path)
			io.WriteString(w, "File "+df.name+" is deleted")
			return nil

		}
	} else {
		//if this file refer to other file, delete only meta data file
		original, err := LoadMetaData(df.refPath)
		if err != nil {
			fmt.Println("file " + df.path + " " + err.Error())
			return err
		}
		//decrease numofrefer
		original.numOfRefer--
		original.WriteMetaData()

		for k := 0; k < len(hash.Referpaths); k++ {
			if hash.Referpaths[k] == df.path {
				//delete k
				newReferPaths := append(hash.Referpaths[:k], hash.Referpaths[k+1:]...)
				hash.Referpaths = newReferPaths
				if err := hashMongo.UpdateHash(hash); err != nil {
					http.Error(w, "DB connection Error", http.StatusInternalServerError)
					log.Fatal("Update hash " + hash.Key + " error: " + err.Error())
					return err
				}
				break
			}
		}

		DeleteFile(df.path + ".meta")
		fmt.Println("file delte OK " + df.path)
		io.WriteString(w, "File "+df.name+" is deleted")

		return nil
	}
	return nil
}
