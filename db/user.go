package db

import (
	"fmt"

	mydb "github.com/bladewaltz9/FileStoreServer/db/mysql"
)

// UserSignup: sign up by username and password
func UserSignup(username string, passwd string) bool {
	insertQuery := "insert ignore into tbl_user(user_name, user_pwd) values(?, ?)"
	stmt, err := mydb.DBConn().Prepare(insertQuery)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	defer stmt.Close()

	ret, err := stmt.Exec(username, passwd)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}

	rowsAffected, err := ret.RowsAffected()
	if err != nil || rowsAffected <= 0 {
		return false
	}
	return true

}

// UserSignin: check username and password
func UserSignin(username string, encPasswd string) bool {
	selectQuery := "select * from tbl_user where user_name = ? limit 1"
	stmt, err := mydb.DBConn().Prepare(selectQuery)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	defer stmt.Close()

	rows, err := stmt.Query(username)
	if err != nil {
		fmt.Println(err.Error())
		return false
	} else if rows == nil {
		fmt.Println("user not found: " + username)
		return false
	}
	pRows := mydb.ParseRows(rows)
	if len(pRows) > 0 && string(pRows[0]["user_pwd"].([]byte)) == encPasswd {
		return true
	}
	return false
}

// UpdateToken: refresh user's login token
func UpdateToken(username string, token string) bool {
	replaceQuery := "replace into tbl_user_token(user_name, user_token) values(?, ?)"
	stmt, err := mydb.DBConn().Prepare(replaceQuery)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	defer stmt.Close()

	_, err = stmt.Exec(username, token)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	return true
}

// GetToken: get token by username from database
func GetToken(username string) (string, error) {
	selectQuery := "select user_token from tbl_user_token where user_name=? limit 1"
	stmt, err := mydb.DBConn().Prepare(selectQuery)
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}
	defer stmt.Close()

	var token string
	err = stmt.QueryRow(username).Scan(&token)
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}
	return token, nil
}

type User struct {
	Username     string
	Email        string
	Phone        string
	SignupAt     string
	LastActiveAt string
	Status       int
}

// GetUserInfo: get user info from database
func GetUserInfo(username string) (User, error) {
	user := User{}
	selectQuery := "select user_name, signup_at from tbl_user where user_name=? limit 1"
	stmt, err := mydb.DBConn().Prepare(selectQuery)
	if err != nil {
		fmt.Println(err.Error())
		return User{}, err
	}
	defer stmt.Close()

	err = stmt.QueryRow(username).Scan(&user.Username, &user.SignupAt)
	if err != nil {
		fmt.Println(err.Error())
		return User{}, err
	}
	return user, nil
}
