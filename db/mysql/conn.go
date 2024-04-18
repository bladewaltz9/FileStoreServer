package mysql

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func init() {
	var err error
	db, err = sql.Open("mysql", "root:123456@tcp(127.0.0.1:3306)/fileserver?charset=utf8")
	if err != nil {
		fmt.Println("Error connnecting to the database: ", err)
		os.Exit(1)
	}

	db.SetMaxOpenConns(1000)

	err = db.Ping()
	if err != nil {
		fmt.Println("Error pinging the database: ", err)
		os.Exit(1)
	}
}

// DBConn: return the object of database connection
func DBConn() *sql.DB {
	return db
}
