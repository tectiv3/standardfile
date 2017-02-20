package db

import (
	"database/sql"
	"log"
	//importing init from sqlite3
	_ "github.com/mattn/go-sqlite3"
)

//Database encapsulates database
type Database struct {
	db *sql.DB
}

//YYYY-MM-DD HH:MM:SS

func (db Database) begin() (tx *sql.Tx) {
	tx, err := db.db.Begin()
	if err != nil {
		log.Println(err)
		return nil
	}
	return tx
}

func (db Database) prepare(q string) (stmt *sql.Stmt) {
	stmt, err := db.db.Prepare(q)
	if err != nil {
		log.Println(err)
		return nil
	}
	return stmt
}

func (db Database) query(q string, args ...interface{}) (rows *sql.Rows) {
	rows, err := db.db.Query(q, args...)
	if err != nil {
		log.Println(err)
		return nil
	}
	return rows
}

var database Database
var err error

func init() {
	database.db, err = sql.Open("sqlite3", "sf.sqlite?parseTime=true")
	if err != nil {
		log.Fatal(err)
	}
}

//Query db function
func Query(sql string, args ...interface{}) error {
	log.Println(sql)
	SQL := database.prepare(sql)
	tx := database.begin()
	_, err = tx.Stmt(SQL).Exec(args...)
	if err != nil {
		log.Println("Query: ", err)
		tx.Rollback()
	} else {
		err = tx.Commit()
		if err != nil {
			log.Println(err)
			return err
		}
		log.Println("Commit successful")
	}
	return err
}
