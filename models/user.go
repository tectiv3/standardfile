package models

import (
	"crypto/sha1"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/satori/go.uuid"
	"github.com/tectiv3/standardfile/db"
)

/*
User is the user type
*/
type User struct {
	Uuid        string `json:"uuid"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	Pw_func     string `json:"pw_func"`
	Pw_alg      string `json:"pw_alg"`
	Pw_cost     int    `json:"pw_cost"`
	Pw_key_size int    `json:"pw_key_size"`
	pw_nonce    string
	Created_at  time.Time `json:"created_at"`
	Updated_at  time.Time `json:"updated_at"`
}

//Save - save current user into DB
func (u *User) Save() {
	u.Uuid = uuid.NewV4().String()
	u.Created_at = time.Now()
	err := db.Query("insert into user(email, password, pw_func, pw_alg, pw_cost, pw_key_size, pw_nonce, created_at) values(?,?,?,?,?,?,?,?)", u.Email, u.Password, u.Pw_func, u.Pw_alg, u.Pw_cost, u.Pw_key_size, u.pw_nonce, u.Created_at)
	if err != nil {
		log.Println(err)
	}
}

//Update - update password
func (u *User) Update(password string) {
	u.Password = password
	u.Updated_at = time.Now()
}

//Exists - checks if current user exists in DB
func (u User) Exists() bool {
	return false
}

//Login - logins user
func (u User) Login() bool {
	return true
}

func (u User) loadByEmail(email string) {

}

//GetParams returns auth parameters by email
func (u User) GetParams(email string) interface{} {
	u.loadByEmail(email)
	params := map[string]string{}
	if u.Pw_cost != 0 {
		params["pw_cost"] = string(u.Pw_cost)
	} else {
		params["pw_cost"] = "5000"
	}
	if u.Pw_alg != "" {
		params["pw_alg"] = u.Pw_alg
	} else {
		params["pw_alg"] = "sha512"
	}
	if u.Pw_key_size != 0 {
		params["pw_key_size"] = string(u.Pw_key_size)
	} else {
		params["pw_key_size"] = "512"
	}
	if u.Pw_func != "" {
		params["pw_func"] = u.Pw_func
	} else {
		params["pw_func"] = "pbkdf2"
	}
	var salt string
	if u.pw_nonce != "" {
		salt = email + "SN" + u.pw_nonce
	} else {
		salt = email + "SN" + "a04a8fe6bcb19ba61c5c0873d391e987982fbbd4"
	}
	params["pw_salt"] = strings.Replace(fmt.Sprintf("% x", sha1.Sum([]byte(salt))), " ", "", -1)

	return params
}
