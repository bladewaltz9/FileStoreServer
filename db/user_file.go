package db

import (
	"fmt"

	mydb "github.com/bladewaltz9/FileStoreServer/db/mysql"
)

type UserFile struct {
	UserName    string
	FileHash    string
	FileName    string
	FileSize    int64
	UploadAt    string
	LastUpdated string
}

// OnUserFileUploadFinished: update user file table
func OnUserFileUploadFinished(username string, fileHash string, fileName string, fileSize int64) bool {
	insertQuery := "insert ignore into tbl_user_file (user_name, file_sha1, file_name, file_size) values(?,?,?,?)"
	stmt, err := mydb.DBConn().Prepare(insertQuery)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	defer stmt.Close()

	_, err = stmt.Exec(username, fileHash, fileName, fileSize)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	return true
}

// GetUserFileMetas: get user files info from database
func GetUserFileMetas(username string, limit int) ([]UserFile, error) {
	selectQuery := "select file_sha1, file_name, file_size, upload_at, last_update from tbl_user_file where user_name=? limit ?"
	stmt, err := mydb.DBConn().Prepare(selectQuery)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(username, limit)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	defer rows.Close()

	var userFiles []UserFile
	for rows.Next() {
		userFile := UserFile{}
		err := rows.Scan(&userFile.FileHash, &userFile.FileName, &userFile.FileSize, &userFile.UploadAt, &userFile.LastUpdated)
		if err != nil {
			fmt.Println(err.Error())
			break
		}

		userFiles = append(userFiles, userFile)
	}
	return userFiles, nil
}
