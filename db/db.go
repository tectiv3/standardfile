package db

import (
	"database/sql"
	"io/ioutil"
	"log"
	"reflect"

	"github.com/kisielk/sqlstruct"
	//importing init from sqlite3
	_ "github.com/mattn/go-sqlite3"
)

//Database encapsulates database
type Database struct {
	db *sql.DB
}

func (db Database) begin() (tx *sql.Tx) {
	tx, err := db.db.Begin()
	if err != nil {
		log.Println(err)
		return nil
	}
	return tx
}

func (db Database) prepare(q string) (stmt *sql.Stmt) {
	// log.Println("Query:", q)
	stmt, err := db.db.Prepare(q)
	if err != nil {
		log.Println(err)
		return nil
	}
	return stmt
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
	stmt := database.prepare(sql)
	defer stmt.Close()
	tx := database.begin()
	if _, err := tx.Stmt(stmt).Exec(args...); err != nil {
		log.Println("Query error: ", err)
		tx.Rollback()
	}
	err := tx.Commit()
	return err
}

//SelectFirst - selects first result from a row
func SelectFirst(sql string, args ...interface{}) (interface{}, error) {
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
	destv := reflect.ValueOf(obj)
	elem := destv.Elem()
	typeOfObj := elem.Type()

	var values []interface{}
	for i := 0; i < elem.NumField(); i++ {
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

//Select - selects multiple results from the DB
func Select(sql string, obj interface{}, args ...interface{}) (result []interface{}, err error) {
	destv := reflect.ValueOf(obj)
	elem := destv.Elem()
	typeOfObj := elem.Type()
	stmt := database.prepare(sql)
	defer stmt.Close()

	rows, err := stmt.Query(args...)
	defer rows.Close()

	for rows.Next() {
		var o = reflect.New(typeOfObj).Interface()
		err = sqlstruct.Scan(o, rows)
		result = append(result, o)
	}
	return result, err
}
