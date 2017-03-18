package main

import (
	"log"

	"github.com/takama/router"
)

func logger(handle router.Handle) router.Handle {
	return func(c *router.Control) {
		log.Printf("%s\t%s", c.Request.Method, c.Request.RequestURI)
		handle(c)
	}
}

func panicHandler(c *router.Control, err interface{}) {
	log.Println(err)
	c.Code(500).Body("")
}

func main() {
	r := router.New()
	r.CustomHandler = logger
	r.PanicHandler = panicHandler

	r.GET("/", Dashboard)

	r.POST("/api/items/sync", SyncItems)
	r.POST("/api/items/backup", BackupItems)
	// r.DELETE("/api/items", DeleteItems)

	r.POST("/api/auth", Registration)
	r.PATCH("/api/auth", ChangePassword)
	r.POST("/api/auth/change_pw", ChangePassword)
	r.POST("/api/auth/sign_in", Login)
	r.POST("/api/auth/sign_in.json", Login)
	r.GET("/api/auth/params", GetParams)

	log.Println("Running on port 8888")
	r.Listen(":8888")
}
