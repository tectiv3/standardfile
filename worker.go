package main

import (
	"github.com/go-playground/pure"
	mw "github.com/go-playground/pure/_examples/middleware/logging-recovery"
	"github.com/tectiv3/standardfile/db"
	"log"
	"net/http"
	"os"
	"strconv"
)

func worker(port int, dbpath string) {
	if port == 0 {
		port = 8888
	}
	db.Init(dbpath)
	log.Println("Started StandardFile Server")
	r := pure.New()
	r.Use(mw.LoggingAndRecovery(true))

	r.Get("/", Dashboard)
	r.Post("/api/items/sync", SyncItems)
	r.Post("/api/items/backup", BackupItems)
	// r.DELETE("/api/items", DeleteItems)
	r.Post("/api/auth", Registration)
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
