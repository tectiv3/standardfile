package db

import (
	"database/sql"
	"io/ioutil"
	"log"
	"reflect"

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

func (db Database) createTables() {
	// create table if not exists
	table, err := ioutil.ReadFile("./db/schema.sql")
	if err != nil {
		panic(err)
	}
	_, err = db.db.Exec(string(table))
	if err != nil {
		panic(err)
	}
}

//eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1dWlkIjoiZDBkYzNjNzQtZTVhZS00NzZhLTkwMWEtZDQzM2Q3OWViNGM5IiwicHdfaGFzaCI6IjZSa0NNNFYwY0ppVUpFUk91IiwiaWF0IjoxNDg3NjExODUzfQ.OH83iy_bim_TNrXCCpdgBhooHhmzosEqflou5Ju1blA
var database Database
var err error

func init() {
	database.db, err = sql.Open("sqlite3", "db/sf.db?loc=auto&parseTime=true")
	// database.db, err = sql.Open("mysql", "Username:Password@tcp(Host:Port)/standardfile?parseTime=true")

	if err != nil {
		log.Fatal(err)
	}
	if database.db == nil {
		log.Fatal("db nil")
	}
	database.createTables()
}

//Query db function
func Query(sql string, args ...interface{}) error {
	log.Println("Query:", sql)
	stmt := database.prepare(sql)
	defer stmt.Close()
	tx := database.begin()
	_, err = tx.Stmt(stmt).Exec(args...)
	if err != nil {
		log.Println("Query: ", err)
		tx.Rollback()
	} else {
		err = tx.Commit()
		if err != nil {
			log.Println("Commit error:", err)
			return err
		}
		log.Println("Commit successful")
	}
	return err
}

//SelectFirst - selects first result from a row
func SelectFirst(sql string, args ...interface{}) (interface{}, error) {
	log.Println("Query:", sql, args)
	stmt := database.prepare(sql)
	defer stmt.Close()
	var result string
	err = stmt.QueryRow(args...).Scan(&result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

//SelectStruct - returns selected result as struct
func SelectStruct(sql string, obj interface{}, args ...interface{}) (interface{}, error) {
	log.Println("Query:", sql, args)
	destv := reflect.ValueOf(obj)
	elem := destv.Elem()
	typeOfObj := elem.Type()

	var values []interface{}
	for i := 0; i < elem.NumField(); i++ {
		// log.Println(typeOfObj.Field(i).Name)
		values = append(values, elem.FieldByName(typeOfObj.Field(i).Name).Addr().Interface())
	}

	stmt := database.prepare(sql)
	defer stmt.Close()
	err := stmt.QueryRow(args...).Scan(values...)
	if err != nil {
		return nil, err
	}
	return obj, nil
}
