package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	dblayer "github.com/bladewaltz9/FileStoreServer/db"
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

		// TODO: 计算大文件的 hash 会比较久，可以进行异步处理
		newFile.Seek(0, 0) // move the file pointer to the beginning of the file
		fileMeta.FileSha1 = util.FileSha1(newFile)
		// meta.UpdateFileMeta(fileMeta)
		_ = meta.UpdateFileMetaDB(fileMeta)

		// update user file table
		r.ParseForm()
		username := r.Form.Get("username")
		suc := dblayer.OnUserFileUploadFinished(username, fileMeta.FileSha1, fileMeta.FileName, fileMeta.FileSize)
		if suc {
			http.Redirect(w, r, "/static/view/home.html", http.StatusFound)
		} else {
			w.Write([]byte("Upload Failed"))
		}
	}
}

// UploadSuccessHandler: file upload finished
func UploadSuccessHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Upload finished!")
}

// GetFileMetaHandler: get file meta information
// TODO: check if the file exists before returning
func GetFileMetaHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fileHash := r.Form["filehash"][0]
	// fileMeta := meta.GetFileMeta(fileHash)
	fileMeta, err := meta.GetFileMetaDB(fileHash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(fileMeta)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

// FileQueryHandler: batch query user's file metas
func FileQueryHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.Form.Get("username")
	limitCnt, _ := strconv.Atoi(r.Form.Get("limit"))

	userFiles, err := dblayer.GetUserFileMetas(username, limitCnt)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(userFiles)
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

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Description", "attachment; filename=\""+fileMeta.FileName+"\"")
	w.Write(data)
}

// FileMetaUpdateHandler: update file meta information(rename)
func FileMetaUpdateHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	opType := r.Form.Get("op")
	fileHash := r.Form.Get("filehash")
	newFileName := r.Form.Get("filename")

	if opType != "0" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	fileMeta := meta.GetFileMeta(fileHash)
	fileMeta.FileName = newFileName
	meta.UpdateFileMeta(fileMeta)

	data, err := json.Marshal(fileMeta)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

// FileDeleteHandler: delete file and meta
func FileDeleteHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fileHash := r.Form.Get("filehash")

	// remove the real file from storage
	fileMeta := meta.GetFileMeta(fileHash)
	os.Remove(fileMeta.Location)

	meta.RemoveFileMeta(fileHash)

	w.WriteHeader(http.StatusOK)
}

func TryFastUploadHandler(w http.ResponseWriter, r *http.Request) {
	// parse request parameters
	r.ParseForm()
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filename := r.Form.Get("filename")
	filesize, _ := strconv.Atoi(r.Form.Get("filesize"))

	// search the file with same hash in tbl_file
	fileMeta, err := meta.GetFileMetaDB(filehash)
	if err != nil {
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// if no record is found, return failure
	if fileMeta == nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "fast upload failed, please visit the normal upload interface.",
		}
		w.Write(resp.JSONBytes())
		return
	} else { // write file info to tbl_user_file, return success
		suc := dblayer.OnUserFileUploadFinished(username, filehash, filename, int64(filesize))
		if suc {
			resp := util.RespMsg{
				Code: 0,
				Msg:  "fast upload successfully.",
			}
			w.Write(resp.JSONBytes())
			return
		} else {
			resp := util.RespMsg{
				Code: -2,
				Msg:  "fast upload failed, please try again later.",
			}
			w.Write(resp.JSONBytes())
			return
		}
	}
}
