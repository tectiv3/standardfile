package db

import (
	"database/sql"
	"log"
	"reflect"

	"github.com/kisielk/sqlstruct"
	//importing init from sqlite3
	_ "github.com/mattn/go-sqlite3"
)

const SCHEMA string = `
PRAGMA foreign_keys=OFF;
BEGIN TRANSACTION;
CREATE TABLE IF NOT EXISTS "items" (
    "uuid" varchar(36) primary key NULL,
    "user_uuid" varchar(36) NOT NULL,
    "content" blob NOT NULL,
    "content_type" varchar(255) NOT NULL,
    "enc_item_key" varchar(255) NOT NULL,
    "auth_hash" varchar(255) NOT NULL,
    "deleted" integer(1) NOT NULL DEFAULT 0,
    "created_at" timestamp NOT NULL,
    "updated_at" timestamp NOT NULL);
CREATE TABLE IF NOT EXISTS "users" ("uuid" varchar(36) primary key NULL, "email" varchar(255) NOT NULL, "password" varchar(255) NOT NULL, "pw_func" varchar(255) NOT NULL DEFAULT "pbkdf2", "pw_alg" varchar(255) NOT NULL DEFAULT "sha512", "pw_cost" integer NOT NULL DEFAULT 5000, "pw_key_size" integer NOT NULL DEFAULT 512, "pw_nonce" varchar(255) NOT NULL, "created_at" timestamp NOT NULL, "updated_at" timestamp NOT NULL);
CREATE INDEX IF NOT EXISTS user_uuid ON items (user_uuid);
CREATE INDEX IF NOT EXISTS user_content on items (user_uuid, content_type);
CREATE INDEX IF NOT EXISTS updated_at on items (updated_at);
CREATE INDEX IF NOT EXISTS email on users (email);
COMMIT;
`

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
	_, err = db.db.Exec(SCHEMA)
	if err != nil {
		panic(err)
	}
}

var database Database
var err error

func Init(dbpath string) {
	database.db, err = sql.Open("sqlite3", dbpath+"?loc=auto&parseTime=true")
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
