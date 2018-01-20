package main

import (
	"crypto/sha1"
	"crypto/sha256"
	"fmt"
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
	Pw_auth     string    `json:"pw_auth"`
	Pw_salt     string    `json:"pw_salt"`
	Created_at  time.Time `json:"created_at"`
	Updated_at  time.Time `json:"updated_at"`
}

//NewPassword - incomming json password change
type NewPassword struct {
	User
	New_password string `json:"new_password"`
}

//UserClaims - jwt claims
type UserClaims struct {
	Uuid    string `json:"uuid"`
	Pw_hash string `json:"pw_hash"`
	jwt.StandardClaims
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
	case "pw_auth":
		u.Pw_auth = value[0]
	case "pw_salt":
		u.Pw_salt = value[0]
	case "pw_cost":
		u.Pw_cost, _ = strconv.Atoi(value[0])
	case "pw_key_size":
		u.Pw_key_size, _ = strconv.Atoi(value[0])
	case "pw_nonce":
		u.Pw_nonce = value[0]
	}
}

//save - save current user into DB
func (u *User) create() error {
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

	err := db.Query("INSERT INTO users (uuid, email, password, pw_func, pw_alg, pw_cost, pw_key_size, pw_nonce, pw_auth, pw_salt, created_at, updated_at) VALUES(?,?,?,?,?,?,?,?,?,?)", u.Uuid, u.Email, u.Password, u.Pw_func, u.Pw_alg, u.Pw_cost, u.Pw_key_size, u.Pw_nonce, u.Pw_auth, u.Pw_salt, u.Created_at, u.Updated_at)

	if err != nil {
		Log(err)
		return err
	}

	return nil
}

//Update - update password
func (u *User) Update(np NewPassword) error {
	if u.Uuid == "" {
		return fmt.Errorf("Unknown user")
	}

	u.Password = Hash(np.New_password)
	u.Pw_func = np.Pw_func
	u.Pw_alg = np.Pw_alg
	u.Pw_cost = np.Pw_cost
	u.Pw_key_size = np.Pw_key_size
	u.Pw_nonce = np.Pw_nonce

	u.Updated_at = time.Now()
	// TODO: validate incomming pw params
	err := db.Query("UPDATE `users` SET `password`=?, `pw_func`=?, `pw_alg`=?, `pw_cost`=?, `pw_key_size`=?, `pw_nonce`=?, `updated_at`=? WHERE `uuid`=?", u.Password, u.Pw_func, u.Pw_alg, u.Pw_cost, u.Pw_key_size, u.Pw_nonce, u.Updated_at, u.Uuid)

	if err != nil {
		Log(err)
		return err
	}

	return nil
}

//Register - creates user and returns token
func (u *User) Register() (string, error) {
	err := u.create()
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
		Log(err)
		return false
	}

	return uuid != ""
}

//Login - logins user
func (u *User) Login(email, password string) (string, error) {
	u.loadByEmailAndPassword(email, Hash(password))

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
		Log(err)
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
		Log(err)
	}
}

func (u *User) loadByEmailAndPassword(email, password string) {
	_, err := db.SelectStruct("SELECT * FROM `users` WHERE `email`=? AND `password`=?", u, email, password)
	if err != nil {
		Log(err)
	}
}

//GetParams returns auth parameters by email
func (u User) GetParams(email string) interface{} {
	u.loadByEmail(email)

	params := map[string]interface{}{}
	params["version"] = "001"
	params["pw_cost"] = u.Pw_cost
	if u.Pw_func != "" {
		params["pw_func"] = u.Pw_func
		params["pw_alg"] = u.Pw_alg
		params["pw_key_size"] = u.Pw_key_size
	}

	if u.Pw_salt == "" {
		nonce := u.Pw_nonce
		if nonce == "" {
			nonce = "a04a8fe6bcb19ba61c5c0873d391e987982fbbd4"
		}
		u.Pw_salt = getSalt(u.Email, nonce)
	}
	params["pw_salt"] = u.Pw_salt

	return params
}

func getSalt(email, nonce string) string {
	return strings.Replace(fmt.Sprintf("% x", sha1.Sum([]byte(email+"SN"+nonce))), " ", "", -1)
}

//Validate - validates password from jwt
func (u User) Validate(password string) bool {
	return password == u.Password
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
