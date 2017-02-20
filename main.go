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

	r.POST("/items/sync", routing.PostSyncFunc)
	r.GET("/auth/params", routing.GetParamsFunc)
	r.POST("/auth/sign_in", routing.PostLoginFunc)
	r.POST("/auth", routing.PostRegisterFunc)
	r.PATCH("/auth", routing.ChangePassFunc)
	r.GET("/", routing.HandleRootFunc)

	log.Print("Running on port 8888")
	r.Listen(":8888")
}
