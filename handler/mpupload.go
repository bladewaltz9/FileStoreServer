package handler

import (
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path"
	"strconv"
	"time"

	rPool "github.com/bladewaltz9/FileStoreServer/cache/redis"
	dblayer "github.com/bladewaltz9/FileStoreServer/db"
	"github.com/bladewaltz9/FileStoreServer/util"
)

// MultipartUploadInfo: multipart upload initialization information
type MultipartUploadInfo struct {
	FileHash   string
	FileSize   int
	UploadID   string
	ChunkSize  int
	ChunkCount int
}

// InitailMultipartUploadHandler: initialize a multipart upload request
func InitailMultipartUploadHandler(w http.ResponseWriter, r *http.Request) {
	// parse request parameters
	r.ParseForm()
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filesize, err := strconv.Atoi(r.Form.Get("filesize"))
	if err != nil {
		w.Write(util.NewRespMsg(-1, "Invalid request parameters", nil).JSONBytes())
		return
	}

	// get a connetion from redis pool
	rConn := rPool.RedisPool()
	// It's unneccessery to close redis in go-redis, because it will be closed automatically

	// generate initialization information for the multipart upload
	upInfo := MultipartUploadInfo{
		FileHash:   filehash,
		FileSize:   filesize,
		UploadID:   username + fmt.Sprintf("%x", time.Now().UnixNano()),
		ChunkSize:  5 * 1024 * 1024, // 5MB
		ChunkCount: int(math.Ceil(float64(filesize) / (5 * 1024 * 1024))),
	}
	// write the initialization information to Redis
	fields := map[string]interface{}{
		"filehash":   upInfo.FileHash,
		"filesize":   upInfo.FileSize,
		"chunkcount": upInfo.ChunkCount,
	}
	rConn.HMSet(rConn.Context(), upInfo.UploadID, fields)

	// return the initialization information to the client
	w.Write(util.NewRespMsg(0, "OK", upInfo).JSONBytes())
}

// UploadPartHandler: handle the upload of a part of the file
func UploadPartHandler(w http.ResponseWriter, r *http.Request) {
	// parse request parameters
	r.ParseForm()
	uploadID := r.Form.Get("uploadid")
	chunkIndex := r.Form.Get("index")

	// get a connetion from redis pool
	rConn := rPool.RedisPool()

	// create a file to store the uploaded part
	filePath := "/home/bladewaltz/data/" + uploadID + "/" + chunkIndex
	err := os.MkdirAll(path.Dir(filePath), 0744)
	if err != nil {
		log.Println("Failed to create file: ", err)
		return
	}
	file, err := os.Create(filePath)
	if err != nil {
		log.Println("Failed to create file: ", err)
		w.Write(util.NewRespMsg(-1, "Upload part failed", nil).JSONBytes())
		return
	}
	defer file.Close()

	// TODO: 对每一个分块做 hash 校验，确保传输过程中数据不被篡改

	// read the file data from the request body
	fileData, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("Failed to read file data: ", err)
		w.Write(util.NewRespMsg(-2, "Upload part failed", nil).JSONBytes())
		return
	}

	// write the file data to the file
	_, err = file.Write(fileData)
	if err != nil {
		log.Println("Failed to write file: ", err)
		w.Write(util.NewRespMsg(-3, "Upload part failed", nil).JSONBytes())
		return
	}

	// update the upload status in Redis
	rConn.HSet(rConn.Context(), "MP_"+uploadID, "chkidx_"+chunkIndex, 1)

	// return success to the client
	w.Write(util.NewRespMsg(0, "OK", nil).JSONBytes())
}

// CompleteUploadHandler: complete the upload of the file
func CompleteUploadHandler(w http.ResponseWriter, r *http.Request) {
	// parse request parameters
	r.ParseForm()
	uploadID := r.Form.Get("uploadid")
	username := r.Form.Get("username")
	fileHash := r.Form.Get("filehash")
	fileSize := r.Form.Get("filesize")
	fileName := r.Form.Get("filename")

	// get a connetion from redis pool
	rConn := rPool.RedisPool()

	// check if the upload is complete
	data, err := rConn.HGetAll(rConn.Context(), "MP_"+uploadID).Result()
	if err != nil {
		log.Println("Failed to get data from Redis: ", err)
		w.Write(util.NewRespMsg(-1, "Complete upload failed", nil).JSONBytes())
		return
	}

	totalCount, _ := strconv.Atoi(data["chunkcount"])
	chunkCount := 0
	for i := 0; i < totalCount; i++ {
		if data["chkidx_"+strconv.Itoa(i)] == "1" {
			chunkCount++
		}
	}

	if chunkCount != totalCount {
		w.Write(util.NewRespMsg(-2, "Some chunks are not uploaded yet", nil).JSONBytes())
		return
	}

	// TODO: merge the parts of the file

	// update the file table and user file table
	fsize, _ := strconv.Atoi(fileSize)
	dblayer.OnFileUploadFinished(fileHash, fileName, int64(fsize), "")
	dblayer.OnUserFileUploadFinished(username, fileHash, fileName, int64(fsize))

	// return success to the client
	w.Write(util.NewRespMsg(0, "OK", nil).JSONBytes())
}
