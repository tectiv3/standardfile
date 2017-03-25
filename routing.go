package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/go-playground/pure"
)

type data map[string]interface{}

func Log(v ...interface{}) {
	if DEBUG {
		log.Println(v...)
	}
}

func showError(w http.ResponseWriter, err error, code int) {
	log.Println(err)
	pure.JSON(w, code, data{"errors": []string{err.Error()}})
}

func authenticateUser(r *http.Request) (User, error) {
	var user = NewUser()

	authHeaderParts := strings.Split(r.Header.Get("Authorization"), " ")
	if len(authHeaderParts) != 2 || strings.ToLower(authHeaderParts[0]) != "bearer" {
		return user, fmt.Errorf("Missing authorization header")
	}

	token, err := jwt.ParseWithClaims(authHeaderParts[1], &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return SigningKey, nil
	})

	if err != nil {
		return user, err
	}

	if claims, ok := token.Claims.(*UserClaims); ok && token.Valid {
		Log("Token is valid, claims: ", claims)
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

//Dashboard - is the root handler
func Dashboard(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Dashboard"))
}

//ChangePassword - is the change password handler
func ChangePassword(w http.ResponseWriter, r *http.Request) {
	user, err := authenticateUser(r)
	if err != nil {
		showError(w, err, http.StatusUnauthorized)
		return
	}
	np := NewPassword{}
	if err := pure.Decode(r, true, 104857600, &np); err != nil {
		showError(w, err, http.StatusUnprocessableEntity)
		return
	}
	Log("Request:", user)

	if err := user.Update(np); err != nil {
		showError(w, err, http.StatusInternalServerError)
		return
	}
	// c.Code(http.StatusNoContent).Body("") //in spec
	token, err := user.Login(user.Email, user.Password)
	if err != nil {
		showError(w, err, http.StatusUnauthorized)
		return
	}
	pure.JSON(w, http.StatusAccepted, data{"token": token, "user": user.ToJSON()})
}

//Registration - is the registration handler
func Registration(w http.ResponseWriter, r *http.Request) {
	var user = NewUser()
	if err := pure.Decode(r, true, 104857600, &user); err != nil {
		showError(w, err, http.StatusUnprocessableEntity)
		return
	}
	Log("Request:", user)
	token, err := user.Register()
	if err != nil {
		showError(w, err, http.StatusUnprocessableEntity)
		return
	}
	pure.JSON(w, http.StatusCreated, data{"token": token, "user": user.ToJSON()})
}

//Login - is the login handler
func Login(w http.ResponseWriter, r *http.Request) {
	var user = NewUser()
	if err := pure.Decode(r, true, 104857600, &user); err != nil {
		showError(w, err, http.StatusUnprocessableEntity)
		return
	}
	Log("Request:", user)
	token, err := user.Login(user.Email, user.Password)
	if err != nil {
		showError(w, err, http.StatusUnauthorized)
		return
	}
	pure.JSON(w, http.StatusAccepted, data{"token": token, "user": user.ToJSON()})
}

//GetParams - is the get auth parameters handler
func GetParams(w http.ResponseWriter, r *http.Request) {
	user := NewUser()
	email := r.FormValue("email")
	Log("Request:", string(email))
	if email == "" {
		showError(w, fmt.Errorf("Empty email"), http.StatusUnauthorized)
		return
	}
	params := user.GetParams(email)
	content, _ := json.MarshalIndent(params, "", "  ")
	Log("Response:", string(content))
	pure.JSON(w, http.StatusOK, params)
}

//SyncItems - is the items sync handler
func SyncItems(w http.ResponseWriter, r *http.Request) {
	user, err := authenticateUser(r)
	if err != nil {
		showError(w, err, http.StatusUnauthorized)
		return
	}
	var request SyncRequest
	if err := pure.Decode(r, true, 104857600, &request); err != nil {
		showError(w, err, http.StatusUnprocessableEntity)
		return
	}
	Log("Request:", request)
	response, err := user.SyncItems(request)
	if err != nil {
		showError(w, err, http.StatusInternalServerError)
		return
	}
	content, _ := json.MarshalIndent(response, "", "  ")
	Log("Response:", string(content))
	pure.JSON(w, http.StatusAccepted, response)
}

//BackupItems - export items
func BackupItems(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		showError(w, err, http.StatusInternalServerError)
		return
	}
	fmt.Printf("%+v\n", r.Form)
}
