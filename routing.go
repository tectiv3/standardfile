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

//Log writes in log if debug flag is set
func Log(v ...interface{}) {
	if *debug {
		log.Println(v...)
	}
}

type sfError struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func showError(w http.ResponseWriter, err error, code int) {
	log.Println(err)
	pure.JSON(w, code, data{"error": sfError{err.Error(), code}})
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

		if ok := user.LoadByUUID(claims.UUID); !ok {
			return user, fmt.Errorf("Unknown user")
		}

		if user.Validate(claims.PwHash) {
			return user, nil
		}
	}

	return user, fmt.Errorf("Invalid token")
}

//Dashboard - is the root handler
func Dashboard(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Dashboard. Server version: " + VERSION))
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

	if len(np.CurrentPassword) == 0 {
		showError(w, fmt.Errorf("Your current password is required to change your password. Please update your application if you do not see this option."), http.StatusUnauthorized)
		return
	}

	if _, err := user.Login(np.Email, np.CurrentPassword); err != nil {
		showError(w, fmt.Errorf("The current password you entered is incorrect. Please try again."), http.StatusUnauthorized)
		return
	}

	if err := user.UpdatePassword(np); err != nil {
		showError(w, err, http.StatusInternalServerError)
		return
	}
	// c.Code(http.StatusNoContent).Body("") //in spec, but SN requires token in return
	token, err := user.Login(user.Email, user.Password)
	if err != nil {
		showError(w, err, http.StatusUnauthorized)
		return
	}
	pure.JSON(w, http.StatusAccepted, data{"token": token, "user": user.ToJSON()})
}

//UpdateUser - updates user params
func UpdateUser(w http.ResponseWriter, r *http.Request) {
	user, err := authenticateUser(r)
	if err != nil {
		showError(w, err, http.StatusUnauthorized)
		return
	}
	p := Params{}
	if err := pure.Decode(r, true, 104857600, &p); err != nil {
		showError(w, err, http.StatusUnprocessableEntity)
		return
	}
	Log("Request:", p)

	if err := user.UpdateParams(p); err != nil {
		showError(w, err, http.StatusInternalServerError)
		return
	}
	pure.JSON(w, http.StatusAccepted, data{})
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
	if _, ok := params["version"]; !ok {
		showError(w, fmt.Errorf("Invalid email or password"), http.StatusNotFound)
		return
	}
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
