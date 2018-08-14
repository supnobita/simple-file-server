# simple-file-server

## Build and Run

- **To build source code and rebuild docker image**
``` bash
cd src code directory (cd simple-file-server)
go get "github.com/gorilla/mux"
go get "gopkg.in/mgo.v2"
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server . 
```

- **Build docker image**
``` bash
git clone https://github.com/supnobita/simple-file-server/
cd simple-file-server
docker build -t simple-file-server:v1 .
```

- **Run container**
```
    cd src code directory (cd simple-file-server)
    docker-compose up -d
```


## API syntax
- Read file: filename.txt:
``` bash
    curl http://127.0.0.1:8080/read?file=filename.txt
```
- Write file: filename.txt:
``` bash
curl --form uploadfile=@filename.txt  http://localhost:8080/upload
```
- Delete file: filename.txt
``` bash
curl http://127.0.0.1:8080/delete?file=filename.txt
```


## Know bug:

=======


## Limitations:
Enormous Limitations !!!!
- single thread
- search file (database)
- deduplication method
- error handler
- foreground job
- write,read speed
