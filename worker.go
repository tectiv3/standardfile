package main

import (
	"github.com/go-playground/pure"
	mw "github.com/go-playground/pure/_examples/middleware/logging-recovery"
	// "github.com/go-playground/pure/middleware"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/tectiv3/standardfile/db"
)

func worker(port int, dbpath string, noreg bool) {
	if port == 0 {
		port = 8888
	}
	db.Init(dbpath)
	log.Println("Started StandardFile Server", VERSION)
	r := pure.New()
	r.Use(mw.LoggingAndRecovery(true))
	// r.Use(mw.LoggingAndRecovery(true), cors)
	// r.RegisterAutomaticOPTIONS(cors)

	r.Get("/", Dashboard)
	r.Post("/api/items/sync", SyncItems)
	r.Post("/api/items/backup", BackupItems)
	// r.DELETE("/api/items", DeleteItems)
	if !noreg {
		r.Post("/api/auth", Registration)
	}
	r.Patch("/api/auth", ChangePassword)
	r.Post("/api/auth/update", UpdateUser)
	r.Post("/api/auth/change_pw", ChangePassword)
	r.Post("/api/auth/sign_in", Login)
	r.Post("/api/auth/sign_in.json", Login)
	r.Get("/api/auth/params", GetParams)

	log.Println("Running on port " + strconv.Itoa(port))
	go listen(r, port)
	<-run
	log.Println("Server stopped")
	os.Exit(0)
}

func listen(r *pure.Mux, port int) {
	err := http.ListenAndServe(":"+strconv.Itoa(port), r.Serve())
	if err != nil {
		log.Println(err)
	}
}

func cors(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "https://app.standardnotes.org")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "authorization,content-type")
		next(w, r)
	}
}
