package models

import (
	"crypto/sha1"
	"crypto/sha256"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/satori/go.uuid"
	"github.com/tectiv3/standardfile/db"
)

//User is the user type
type User struct {
	Uuid        string    `json:"uuid"`
	Email       string    `json:"email"`
	Password    string    `json:"password"`
	Pw_func     string    `json:"pw_func"`
	Pw_alg      string    `json:"pw_alg"`
	Pw_cost     int       `json:"pw_cost"`
	Pw_key_size int       `json:"pw_key_size"`
	Pw_nonce    string    `json:"pw_nonce"`
	Created_at  time.Time `json:"created_at"`
	Updated_at  time.Time `json:"updated_at"`
}

//UserClaims - jwt claims
type UserClaims struct {
	Uuid    string `json:"uuid"`
	Pw_hash string `json:"pw_hash"`
	jwt.StandardClaims
}

type Loadable interface {
	LoadValue(name string, value []string)
}

//SigningKey - export to routing
var SigningKey = []byte{}

func init() {
	key := os.Getenv("SECRET_KEY_BASE")
	if key == "" {
		key = "qA6irmDikU6RkCM4V0cJiUJEROuCsqTa1esexI4aWedSv405v8lw4g1KB1nQVsSdCrcyRlKFdws4XPlsArWwv9y5Xr5Jtkb11w1NxKZabOUa7mxjeENuCs31Y1Ce49XH9kGMPe0ms7iV7e9F6WgnsPFGOlIA3CwfGyr12okas2EsDd71SbSnA0zJYjyxeCVCZJWISmLB"
	}
	SigningKey = []byte(key)
}

//NewUser - user constructor
func NewUser() User {
	user := User{}
	user.Pw_cost = 5000
	user.Pw_alg = "sha512"
	user.Pw_key_size = 512
	user.Pw_func = "pbkdf2"
	user.Created_at = time.Now()
	user.Updated_at = time.Now()
	return user
}

//LoadValue - hydrate struct from map
func (u *User) LoadValue(name string, value []string) {
	switch name {
	case "uuid":
		u.Uuid = value[0]
	case "email":
		u.Email = value[0]
	case "password":
		u.Password = value[0]
	case "pw_func":
		u.Pw_func = value[0]
	case "pw_alg":
		u.Pw_alg = value[0]
	case "pw_cost":
		u.Pw_cost, _ = strconv.Atoi(value[0])
	case "pw_key_size":
		u.Pw_key_size, _ = strconv.Atoi(value[0])
	case "pw_nonce":
		u.Pw_nonce = value[0]
	}
}

//LoadModel - hydrate model
func LoadModel(loadable Loadable, data map[string][]string) {
	for key, value := range data {
		loadable.LoadValue(key, value)
	}
}

//save - save current user into DB
func (u *User) save() error {
	if u.Uuid != "" {
		return fmt.Errorf("Trying to save existing user")
	}

	if u.Email == "" || u.Password == "" {
		return fmt.Errorf("Empty email or password")
	}

	if u.Exists() {
		return fmt.Errorf("Unable to register")
	}

	u.Uuid = uuid.NewV4().String()
	u.Password = Hash(u.Password)
	u.Created_at = time.Now()

	err := db.Query("INSERT INTO users (uuid, email, password, pw_func, pw_alg, pw_cost, pw_key_size, pw_nonce, created_at, updated_at) VALUES(?,?,?,?,?,?,?,?,?,?)", u.Uuid, u.Email, u.Password, u.Pw_func, u.Pw_alg, u.Pw_cost, u.Pw_key_size, u.Pw_nonce, u.Created_at, u.Updated_at)

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

//Update - update password
func (u *User) Update(password string) error {
	if u.Uuid == "" {
		return fmt.Errorf("Unknown user")
	}

	u.Password = Hash(password)
	u.Updated_at = time.Now()

	err := db.Query("UPDATE `users` SET `password`=? WHERE `uuid`=?", u.Password, u.Uuid)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

//Register - creates user and returns token
func (u *User) Register() (string, error) {
	err := u.save()
	if err != nil {
		return "", err
	}

	token, err := u.Login(u.Email, u.Password)
	if err != nil {
		return "", fmt.Errorf("Registration failed")
	}

	return token, nil
}

//Exists - checks if current user exists in DB
func (u User) Exists() bool {
	uuid, err := db.SelectFirst("SELECT `uuid` FROM `users` WHERE `email`=?", u.Email)

	if err != nil {
		log.Println(err)
		return false
	}

	return uuid != ""
}

//Login - logins user
func (u *User) Login(email, password string) (string, error) {
	u.loadByEmailAndPassword(email, password)

	if u.Uuid == "" {
		return "", fmt.Errorf("Invalid email or password")
	}

	token, err := u.CreateToken()
	if err != nil {
		return "", err
	}

	return token, nil
}

//LoadByUUID - loads user info from DB
func (u *User) LoadByUUID(uuid string) bool {
	_, err := db.SelectStruct("SELECT * FROM `users` WHERE `uuid`=?", u, uuid)
	if err != nil {
		log.Println(err)
		return false
	}

	return true
}

//CreateToken - will create JWT token
func (u User) CreateToken() (string, error) {
	claims := UserClaims{
		u.Uuid,
		u.Password,
		jwt.StandardClaims{
			IssuedAt: time.Now().Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(SigningKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (u *User) loadByEmail(email string) {
	_, err := db.SelectStruct("SELECT * FROM `users` WHERE `email`=?", u, email)
	if err != nil {
		log.Println(err)
	}
}

func (u *User) loadByEmailAndPassword(email, password string) {
	_, err := db.SelectStruct("SELECT * FROM `users` WHERE `email`=? AND `password`=?", u, email, password)
	if err != nil {
		log.Println(err)
	}
}

//GetParams returns auth parameters by email
func (u User) GetParams(email string) interface{} {
	// if u.Email != email {}
	// u.loadByEmail(email)

	params := map[string]string{}
	params["pw_cost"] = strconv.Itoa(u.Pw_cost)
	params["pw_alg"] = u.Pw_alg
	params["pw_key_size"] = strconv.Itoa(u.Pw_key_size)
	params["pw_func"] = u.Pw_func

	var salt string
	if u.Pw_nonce != "" {
		salt = email + "SN" + u.Pw_nonce
	} else {
		salt = email + "SN" + "a04a8fe6bcb19ba61c5c0873d391e987982fbbd4"
	}
	params["pw_salt"] = strings.Replace(fmt.Sprintf("% x", sha1.Sum([]byte(salt))), " ", "", -1)

	return params
}

//Validate - validates password from jwt
func (u User) Validate(password string) bool {
	// base64.URLEncoding.EncodeToString()
	pw := Hash(password)
	return pw != u.Password
}

//ToJSON - return map without pw and nonce
func (u User) ToJSON() interface{} {
	u.Password = ""
	u.Pw_nonce = ""
	return u
}

//Hash - sha256 hash function
func Hash(input string) string {
	return strings.Replace(fmt.Sprintf("% x", sha256.Sum256([]byte(input))), " ", "", -1)
}
