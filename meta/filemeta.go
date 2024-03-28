package meta

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

// UpdateFileMeta: add or update file meta information
func UpdateFileMeta(fileMeta FileMeta) {
	fileMetas[fileMeta.FileSha1] = fileMeta
}

// GetFileMeta: get file by the value of Sha1
func GetFileMeta(fileSha1 string) FileMeta {
	return fileMetas[fileSha1]
}

// RemoveFileMeta: delete file meta information
func RemoveFileMeta(fileSha1 string) {
	delete(fileMetas, fileSha1)
}
