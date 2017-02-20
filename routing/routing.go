package routing

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/satori/go.uuid"
	"github.com/takama/router"
	"github.com/tectiv3/standardfile/models"
)

type data map[string]interface{}

var signingKey = []byte{}

func init() {
	key := os.Getenv("SECRET_KEY_BASE")
	if key == "" {
		key = "qA6irmDikU6RkCM4V0cJiUJEROuCsqTa1esexI4aWedSv405v8lw4g1KB1nQVsSdCrcyRlKFdws4XPlsArWwv9y5Xr5Jtkb11w1NxKZabOUa7mxjeENuCs31Y1Ce49XH9kGMPe0ms7iV7e9F6WgnsPFGOlIA3CwfGyr12okas2EsDd71SbSnA0zJYjyxeCVCZJWISmLB"
	}
	signingKey = []byte(key)
}

//HandleRootFunc - is the root handler
func HandleRootFunc(c *router.Control) {
	checkHeader(c)
	c.UseTimer()
	items := models.Items{
		models.Item{Content: "first", Uuid: uuid.NewV4().String()},
		models.Item{Content: "second", Uuid: uuid.NewV4().String()},
	}
	c.Code(200).Body(items)
}

//ChangePassFunc - is the change password handler
func ChangePassFunc(c *router.Control) {
	c.Code(204)
}

//parseRequest - is an internal function to parse json from request into local struct
func parseRequest(c *router.Control, value interface{}) {
	r := c.Request
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		panic(err)
	}
	if err := r.Body.Close(); err != nil {
		panic(err)
	}
	if err := json.Unmarshal(body, &value); err != nil {
		log.Println(err)
		c.Code(422).Body(data{"errors": []string{err.Error()}})
	}
}

//PostRegisterFunc - is the registration handler
func PostRegisterFunc(c *router.Control) {
	var user models.User
	parseRequest(c, &user)
	if user.Exists() {
		c.Code(422).Body(data{"errors": []string{"Unable to register!"}})
	}
	user.Save()
	ok := user.Login()
	if !ok {
		c.Code(422).Body(data{"errors": []string{"Unable to register."}})
	}
	token := CreateToken("")
	c.Code(http.StatusCreated).Body(data{"token": token, "user": user})
}

//PostLoginFunc - is the login handler
func PostLoginFunc(c *router.Control) {
	var user models.User
	parseRequest(c, &user)
	ok := user.Login()
	if !ok {
		c.Code(422).Body(data{"errors": []string{"Invalid email or password."}})
	}
	token := CreateToken("")
	c.Code(http.StatusCreated).Body(data{"token": token, "user": user})
}

//PostSyncFunc - is the items sync handler
func PostSyncFunc(c *router.Control) {
	checkHeader(c)
	var message string
	message = "POST Sync"
	c.Body(message)
}

//GetParamsFunc - is the get auth parameters handler
func GetParamsFunc(c *router.Control) {
	var user models.User
	params := user.GetParams(c.Request.FormValue("email"))
	// parseRequest(c, &user)
	// params := user.getParams()
	c.Code(200).Body(params)
}

// eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.e30.QfgSueR9Q2kaNInrLUdtZwv1hg64epO3aaTFMjJeqms

//CreateToken - will create JWT token
func CreateToken(payload string) string {
	token := jwt.New(jwt.SigningMethodHS256)
	tokenString, err := token.SignedString(signingKey)

	if err != nil {
		log.Println("Error signing token", err)
		return ""
	}

	return tokenString
}

//ValidateToken - will validate the token
func ValidateToken(myToken string) bool {
	token, err := jwt.Parse(myToken, func(token *jwt.Token) (interface{}, error) {
		return []byte(signingKey), nil
	})

	if err != nil {
		log.Println("Error validating token", err)
		return false
	}

	return token.Valid
}

func checkHeader(c *router.Control) (bool, error) {
	authHeaderParts := strings.Split(c.Request.Header.Get("Authorization"), " ")
	if len(authHeaderParts) != 2 || strings.ToLower(authHeaderParts[0]) != "bearer" {
		return false, fmt.Errorf("Authorization header format must be Bearer {token}")
	}
	token := authHeaderParts[1]
	if !ValidateToken(token) {
		c.Code(http.StatusInternalServerError).Body(data{"errors": []string{"Invalid token"}})
	}
	log.Println("Token valid:", token)
	return true, nil
}
