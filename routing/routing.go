package routing

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/satori/go.uuid"
	"github.com/takama/router"
	"github.com/tectiv3/standardfile/models"
)

type data map[string]interface{}

//HandleRootFunc - is the root handler
func HandleRootFunc(c *router.Control) {
	user, err := authenticateUser(c)
	if err != nil {
		log.Println(err)
		c.Code(http.StatusUnauthorized).Body(data{"errors": []string{err.Error()}})
		return
	}
	c.UseTimer()
	items := data{
		"items": models.Items{
			models.Item{Content: "first", Uuid: uuid.NewV4().String()},
			models.Item{Content: "second", Uuid: uuid.NewV4().String()},
		},
		"user": user,
	}
	c.Code(200).Body(items)
}

//ChangePassFunc - is the change password handler
func ChangePassFunc(c *router.Control) {
	c.Code(204)
}

//parseRequest - is an internal function to parse json from request into local struct
func parseRequest(c *router.Control, value interface{}) error {
	r := c.Request
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		return err
	}
	if err := r.Body.Close(); err != nil {
		return err
	}
	if len(body) == 0 {
		return fmt.Errorf("Empty request")
	}
	if err := json.Unmarshal(body, &value); err != nil {
		return err
	}
	return nil
}

//PostRegisterFunc - is the registration handler
func PostRegisterFunc(c *router.Control) {
	var user = models.NewUser()
	err := parseRequest(c, &user)
	if err != nil {
		log.Println(err)
		c.Code(422).Body(data{"errors": []string{err.Error()}})
		return
	}
	if user.Email == "" || user.Password == "" {
		log.Println("Empty email or password")
		c.Code(422).Body(data{"errors": []string{"Unable to register."}})
		return
	}
	if user.Exists() {
		log.Println("Users exists")
		c.Code(422).Body(data{"errors": []string{"Unable to register!"}})
		return
	}
	err = user.Save()
	if err != nil {
		log.Println(err)
		c.Code(422).Body(data{"errors": []string{err.Error()}})
		return
	}
	ok := user.Login()
	if !ok {
		c.Code(422).Body(data{"errors": []string{"Unable to register."}})
		return
	}
	token := user.CreateToken()
	c.Code(http.StatusCreated).Body(data{"token": token, "user": user})
}

//PostLoginFunc - is the login handler
func PostLoginFunc(c *router.Control) {
	var user = models.NewUser()
	parseRequest(c, &user)
	ok := user.Login()
	if !ok {
		c.Code(422).Body(data{"errors": []string{"Invalid email or password."}})
	}
	token := user.CreateToken()
	c.Code(http.StatusAccepted).Body(data{"token": token, "user": user})
}

//PostSyncFunc - is the items sync handler
func PostSyncFunc(c *router.Control) {
	_, err := authenticateUser(c)
	if err != nil {
		log.Println(err)
		c.Code(http.StatusUnauthorized).Body(data{"errors": []string{err.Error()}})
		return
	}
	var message string
	message = "POST Sync"
	c.Body(message)
}

//GetParamsFunc - is the get auth parameters handler
func GetParamsFunc(c *router.Control) {
	var user = models.NewUser()
	email := c.Request.FormValue("email")
	if email == "" {
		log.Println("Empty email")
		c.Code(http.StatusUnauthorized).Body(data{"errors": []string{"Empty email"}})
		return
	}
	params := user.GetParams(email)
	// parseRequest(c, &user)
	// params := user.getParams()
	c.Code(200).Body(params)
}

func authenticateUser(c *router.Control) (models.User, error) {
	var user = models.NewUser()

	authHeaderParts := strings.Split(c.Request.Header.Get("Authorization"), " ")
	if len(authHeaderParts) != 2 || strings.ToLower(authHeaderParts[0]) != "bearer" {
		return user, fmt.Errorf("Authorization header format must be Bearer {token}")
	}

	token, err := jwt.ParseWithClaims(authHeaderParts[1], &models.UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return models.SigningKey, nil
	})

	if err != nil {
		return user, err
	}

	if claims, ok := token.Claims.(*models.UserClaims); ok && token.Valid {
		log.Println("Token is valid, claims: ", claims)
		ok = user.LoadByUUID(claims.Uuid)
		if !ok {
			return user, fmt.Errorf("Unknown user")
		}
		if user.Validate(claims.Pw_hash) {
			return user, nil
		}
		return user, fmt.Errorf("Old password used for authorisation")
	}

	return user, fmt.Errorf("Invalid token")
}
