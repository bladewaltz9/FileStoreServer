package db

import (
	"database/sql"
	"fmt"

	mydb "github.com/bladewaltz9/FileStoreServer/db/mysql"
)

// OnFileUploadFinished: file upload successfully, insert meta info into database
func OnFileUploadFinished(fileHash string, fileName string, fileSize int64, fileAddr string) bool {
	insertQuery := "insert ignore into tbl_file (file_sha1, file_name, file_size, file_addr, status) values (?, ?, ?, ?, 1)"
	stmt, err := mydb.DBConn().Prepare(insertQuery)
	if err != nil {
		fmt.Println("Failed to prepare statement, err: ", err)
		return false
	}
	defer stmt.Close()

	ret, err := stmt.Exec(fileHash, fileName, fileSize, fileAddr)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}

	rowsAffected, err := ret.RowsAffected()
	if rowsAffected <= 0 {
		fmt.Printf("WARNING: File with hash: %s has been uploaded before.\n", fileHash)
	}
	return true
}

type TableFile struct {
	FileHash string
	FileName sql.NullString
	FileSize sql.NullInt64
	FileAddr sql.NullString
}

// GetFileMeta: get file meta info from MySQL by file hash
func GetFileMeta(fileHash string) (*TableFile, error) {
	selectQuery := "select file_sha1, file_name, file_size, file_addr from tbl_file where file_sha1 = ? and status = 1 limit 1"
	stmt, err := mydb.DBConn().Prepare(selectQuery)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	defer stmt.Close()

	tableFile := TableFile{}
	err = stmt.QueryRow(fileHash).Scan(&tableFile.FileHash, &tableFile.FileName, &tableFile.FileSize, &tableFile.FileAddr)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	return &tableFile, nil
}
