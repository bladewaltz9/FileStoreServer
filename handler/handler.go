package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/bladewaltz9/FileStoreServer/meta"
	"github.com/bladewaltz9/FileStoreServer/util"
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

		fileMeta := meta.FileMeta{
			FileName:   header.Filename,
			Location:   "/tmp/" + header.Filename,
			UploadTime: time.Now().Format("2006-01-02 15:04:05"),
		}

		newFile, err := os.Create(fileMeta.Location)
		if err != nil {
			fmt.Println("Failed to create file, err: " + err.Error())
			return
		}
		defer newFile.Close()

		fileMeta.FileSize, err = io.Copy(newFile, file)
		if err != nil {
			fmt.Println("Failed to store data to file, err: " + err.Error())
			return
		}

		newFile.Seek(0, 0) // move the file pointer to the beginning of the file
		fileMeta.FileSha1 = util.FileSha1(newFile)
		meta.UpdateFileMeta(fileMeta)

		fmt.Println(fileMeta)

		http.Redirect(w, r, "/file/upload/success", http.StatusFound)
	}
}

// UploadSuccessHandler: file upload finished
func UploadSuccessHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Upload finished!")
}

// GetFileMetaHandler: get file meta information
func GetFileMetaHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fileHash := r.Form["filehash"][0]
	fileMeta := meta.GetFileMeta(fileHash)
	data, err := json.Marshal(fileMeta)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

// DownloadHandler: handle file download
func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fileHash := r.Form.Get("filehash")
	fileMeta := meta.GetFileMeta(fileHash)

	fmt.Println("filehash: " + fileHash)
	fmt.Println("download: " + fileMeta.FileName)

	file, err := os.Open(fileMeta.Location)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octect-stream")
	w.Header().Set("Content-Description", "attachment; filename=\""+fileMeta.FileName+"\"")
	w.Write(data)
}
