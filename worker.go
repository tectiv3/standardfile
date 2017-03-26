package main

import (
	"github.com/go-playground/pure"
	mw "github.com/go-playground/pure/examples/middleware/logging-recovery"
	"log"
	"net/http"
	"os"
)

func worker() {
	log.Println("Started StandardFile Server")
	r := pure.New()
	r.Use(mw.LoggingAndRecovery(true))

	r.Get("/", Dashboard)
	r.Post("/api/items/sync", SyncItems)
	r.Post("/api/items/backup", BackupItems)
	// r.DELETE("/api/items", DeleteItems)
	r.Post("/api/auth", Registration)
	r.Patch("/api/auth", ChangePassword)
	r.Post("/api/auth/change_pw", ChangePassword)
	r.Post("/api/auth/sign_in", Login)
	r.Post("/api/auth/sign_in.json", Login)
	r.Get("/api/auth/params", GetParams)

	log.Println("Running on port 8888")
	go listen(r)
	<-run
	log.Println("Server stopped")
	os.Exit(0)
}

func listen(r *pure.Mux) {
	err := http.ListenAndServe(":8888", r.Serve())
	if err != nil {
		log.Println(err)
	}
}
