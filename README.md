#simple-file-server

##1. Build and Run
- **Build docker image
...git clone https://github.com/supnobita/simple-file-server/
...cd simple-file-server
...docker build -t simple-file-server:v1 .

- **Run container
...docker run -p 8080:8080 -v /root/data:/go/data simple-file-server:v1

- **To build source code and rebuild docker image
...cd src code directory (cd simple-file-server)
...run cmd ``` CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server . ```

##2. API syntax
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

##3. Know bug:
This code calculate md5 hash of file and save it to .meta file. When you upload file A, B, C has a same content. Assume that file A is upload first, then B and C. If you not delete all B, C file, you can not delete A, and it will return this error:
> File is Original and has many refer
This is stupid bug, I will fix it in furture when i have time.

##4. Limitations:
Enormous Limitations !!!!
- single thread
- search file (database)
- deduplication method
- error handler
- foreground job
- write,read speed
