package main

import (
	"log"

	"github.com/takama/router"
	"github.com/tectiv3/standardfile/routing"
)

func logger(handle router.Handle) router.Handle {
	return func(c *router.Control) {
		log.Printf(
			"%s\t%s",
			c.Request.Method,
			c.Request.RequestURI,
		)
		handle(c)
	}
}

func main() {
	r := router.New()
	r.CustomHandler = logger

	r.GET("/", routing.Dashboard)
	r.POST("/api/items/sync", routing.SyncItems)
	r.GET("/api/auth/params", routing.GetParams)
	r.POST("/api/auth/sign_in", routing.Login)
	r.POST("/api/auth", routing.Registration)
	r.PATCH("/api/auth", routing.ChangePassword)

	log.Print("Running on port 8888")
	r.Listen(":8888")
}
