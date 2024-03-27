package main

import (
	"fmt"
	"net/http"

	"github.com/bladewaltz9/FileStoreServer/handler"
)

func main() {
	http.HandleFunc("/file/upload", handler.UploadHandler)
	http.HandleFunc("/file/upload/success", handler.UploadSuccessHandler)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Failed to start server, err: " + err.Error())
	}
}
