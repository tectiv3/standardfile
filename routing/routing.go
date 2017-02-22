package routing

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/takama/router"
	"github.com/tectiv3/standardfile/models"
)

type data map[string]interface{}

func showError(c *router.Control, err error, code int) {
	log.Println(err)
	c.Code(code).Body(data{"errors": []string{err.Error()}})
}

func authenticateUser(c *router.Control) (models.User, error) {
	var user = models.NewUser()

	authHeaderParts := strings.Split(c.Request.Header.Get("Authorization"), " ")
	if len(authHeaderParts) != 2 || strings.ToLower(authHeaderParts[0]) != "bearer" {
		return user, fmt.Errorf("Missing authorization header")
	}

	token, err := jwt.ParseWithClaims(authHeaderParts[1], &models.UserClaims{}, func(token *jwt.Token) (interface{}, error) {
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

//_parseRequest - is an internal function to parse json from request into local struct
func _parseRequest(c *router.Control, value models.Loadable) error {
	r := c.Request
	ct := r.Header.Get("Content-Type")
	mediatype, _, _ := mime.ParseMediaType(ct)
	if mediatype == "application/json" {
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
		log.Println("Request:", string(body))
		if err := json.Unmarshal(body, &value); err != nil {
			return err
		}
	} else {
		err := r.ParseForm()
		if err != nil {
			return err
		}
		fmt.Printf("%+v\n", r.Form)
		models.Hydrate(value, r.Form)
	}
	return nil
}

//Dashboard - is the root handler
func Dashboard(c *router.Control) {
	c.Code(http.StatusOK).Body("Standard File")
}

//ChangePassword - is the change password handler
func ChangePassword(c *router.Control) {
	user, err := authenticateUser(c)
	if err != nil {
		showError(c, err, http.StatusUnauthorized)
		return
	}
	ct := c.Request.Header.Get("Content-Type")
	mediatype, _, _ := mime.ParseMediaType(ct)
	if mediatype == "application/json" {
		np := models.NewPassword{}
		if err := _parseRequest(c, &np); err != nil {
			showError(c, err, http.StatusUnprocessableEntity)
			return
		}
		if err := user.Update(np); err != nil {
			showError(c, err, http.StatusInternalServerError)
			return
		}
		// c.Code(http.StatusNoContent).Body("") //in spec
		token, err := user.Login(user.Email, user.Password)
		if err != nil {
			showError(c, err, http.StatusUnauthorized)
			return
		}
		c.Code(http.StatusAccepted).Body(data{"token": token, "user": user.ToJSON()})
		return
	}
	//email,new_pw,old_pw
	if err := c.Request.ParseForm(); err != nil {
		showError(c, err, http.StatusUnauthorized)
		return
	}
	fmt.Printf("%+v\n", c.Request.Form)
	log.Println(user)
	c.Code(http.StatusInternalServerError).Body("Unimplemented")
}

//Registration - is the registration handler
func Registration(c *router.Control) {
	var user = models.NewUser()
	if err := _parseRequest(c, &user); err != nil {
		showError(c, err, http.StatusUnprocessableEntity)
		return
	}
	token, err := user.Register()
	if err != nil {
		showError(c, err, http.StatusUnprocessableEntity)
		return
	}
	c.Code(http.StatusCreated).Body(data{"token": token, "user": user.ToJSON()})
}

//Login - is the login handler
func Login(c *router.Control) {
	var user = models.NewUser()
	if err := _parseRequest(c, &user); err != nil {
		showError(c, err, http.StatusUnprocessableEntity)
		return
	}
	token, err := user.Login(user.Email, models.Hash(user.Password))
	if err != nil {
		showError(c, err, http.StatusUnauthorized)
		return
	}
	c.Code(http.StatusAccepted).Body(data{"token": token, "user": user.ToJSON()})
}

//GetParams - is the get auth parameters handler
func GetParams(c *router.Control) {
	user := models.NewUser()
	email := c.Request.FormValue("email")
	if email == "" {
		showError(c, fmt.Errorf("Empty email"), http.StatusUnauthorized)
		return
	}
	params := user.GetParams(email)
	c.Code(200).Body(params)
}

//SyncItems - is the items sync handler
func SyncItems(c *router.Control) {
	user, err := authenticateUser(c)
	if err != nil {
		showError(c, err, http.StatusUnauthorized)
		return
	}
	var request models.SyncRequest
	if e := _parseRequest(c, &request); e != nil {
		showError(c, e, http.StatusUnprocessableEntity)
		return
	}
	response, err := user.SyncItems(request)
	if err != nil {
		showError(c, err, http.StatusInternalServerError)
		return
	}
	content, _ := json.MarshalIndent(response, "", "  ")
	log.Println("Response:", string(content))
	c.Code(http.StatusAccepted).Body(response)
}

//BackupItems - export items
func BackupItems(c *router.Control) {
	err := c.Request.ParseForm()
	if err != nil {
		showError(c, err, http.StatusInternalServerError)
		return
	}
	fmt.Printf("%+v\n", c.Request.Form)
}
