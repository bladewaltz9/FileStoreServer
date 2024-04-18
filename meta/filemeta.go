package meta

import (
	mydb "github.com/bladewaltz9/FileStoreServer/db"
)

// FileMeta: file meta information struct
type FileMeta struct {
	FileSha1   string
	FileName   string
	FileSize   int64
	Location   string
	UploadTime string
}

var fileMetas map[string]FileMeta

func init() {
	fileMetas = make(map[string]FileMeta)
}

// UpdateFileMeta: add/update file meta information
func UpdateFileMeta(fileMeta FileMeta) {
	fileMetas[fileMeta.FileSha1] = fileMeta
}

// UpdateFileMetaDB: add/update file meta info to datebase
func UpdateFileMetaDB(fileMeta FileMeta) bool {
	return mydb.OnFileUploadFinished(fileMeta.FileSha1, fileMeta.FileName, fileMeta.FileSize, fileMeta.Location)
}

// GetFileMeta: get file meta info by Sha1
func GetFileMeta(fileSha1 string) FileMeta {
	return fileMetas[fileSha1]
}

// GetFileMetaDB: gt file meta info from database
func GetFileMetaDB(fileSha1 string) (FileMeta, error) {
	tableFile, err := mydb.GetFileMeta(fileSha1)
	if err != nil {
		return FileMeta{}, err
	}
	fileMeta := FileMeta{
		FileSha1: tableFile.FileHash,
		FileName: tableFile.FileName.String,
		FileSize: tableFile.FileSize.Int64,
		Location: tableFile.FileAddr.String,
	}
	return fileMeta, nil
}

// RemoveFileMeta: delete file meta information
func RemoveFileMeta(fileSha1 string) {
	delete(fileMetas, fileSha1)
}
