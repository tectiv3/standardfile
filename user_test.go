package main_test

import (
	sf "github.com/tectiv3/standardfile"
	"github.com/tectiv3/standardfile/db"
	"testing"
)

var (
	login = sf.User{
		Email:    "user2@local",
		Password: "3cb5561daa49bd5b4438ad214a6f9a6d9b056a2c0b9a91991420ad9d658b8fac",
	}
	register = sf.User{
		Email:       "user2@local",
		Password:    "3cb5561daa49bd5b4438ad214a6f9a6d9b056a2c0b9a91991420ad9d658b8fac",
		Pw_cost:     101000,
		Pw_salt:     "685bdeca99977eb0a30a68284d86bbb322c3b0ee832ffe4b6b76bd10fe7b8362",
		Pw_alg:      "sha512",
		Pw_key_size: 512,
		Pw_func:     "pbkdf2",
	}
)

func init() {
	db.Init(":memory:")
}

func TestRegister(t *testing.T) {
	var user = register
	token, err := user.Register()
	if err != nil {
		t.Error("Register failed", err)
		return
	}
	t.Log("Token:", token)
}

func TestLogin(t *testing.T) {
	var user = login
	token, err := user.Login(user.Email, user.Password)
	if err != nil {
		t.Error("Login failed", err)
		return
	}
	t.Log("Token:", token)
}
