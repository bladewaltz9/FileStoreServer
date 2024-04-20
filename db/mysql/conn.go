package mysql

import (
	"database/sql"
	"fmt"
	"log"
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

// ParseRows: convert SQL query result into slices, each slice represents a row of data
func ParseRows(rows *sql.Rows) []map[string]interface{} {
	columns, _ := rows.Columns()
	scanArgs := make([]interface{}, len(columns)) // store pointer of scan results
	values := make([]interface{}, len(columns))   // store real column value
	for j := range values {
		scanArgs[j] = &values[j]
	}

	record := make(map[string]interface{})
	records := make([]map[string]interface{}, 0)
	for rows.Next() {
		// store row datas into record map
		err := rows.Scan(scanArgs...)
		if err != nil {
			log.Fatal(err)
		}

		for i, col := range values {
			if col != nil {
				record[columns[i]] = col
			}
		}
		records = append(records, record)
	}
	return records
}
