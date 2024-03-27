package handler

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

// UploadHandler: handle file upload
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// return the uploaded HTML page
		data, err := os.ReadFile("./static/view/index.html")
		if err != nil {
			io.WriteString(w, "internel server error")
			return
		}
		io.WriteString(w, string(data))
	} else if r.Method == "POST" {
		// receive file stream and store to local directory
		file, header, err := r.FormFile("file")
		if err != nil {
			fmt.Println("Failed to get data, err: " + err.Error())
			return
		}
		defer file.Close()

		newFile, err := os.Create("/tmp/" + header.Filename)
		if err != nil {
			fmt.Println("Failed to create file, err: " + err.Error())
			return
		}
		defer newFile.Close()

		_, err = io.Copy(newFile, file)
		if err != nil {
			fmt.Println("Failed to store data to file, err: " + err.Error())
			return
		}

		http.Redirect(w, r, "/file/upload/success", http.StatusFound)
	}
}

// UploadSuccessHandler: file upload finished
func UploadSuccessHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Upload finished!")
}
