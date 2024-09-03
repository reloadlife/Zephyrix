package main

import (
	"context"
	"log"

	"net/http"
	_ "net/http/pprof"

	"github.com/gin-gonic/gin"
	"go.mamad.dev/zephyrix"
)

// THIS IS NOT SAFE, NEVER DO IN PROD, THIS IS ONLY HERE TO TEST IF DI WORKS
type routeHandler struct {
	conf *zephyrix.Config
}

func newRouteHandler(conf *zephyrix.Config) *routeHandler {
	return &routeHandler{
		conf: conf,
	}
}

func (r *routeHandler) Method() []string {
	return []string{"GET"}
}

func (r *routeHandler) Path() string {
	return "/hello_world_di"
}

func (r *routeHandler) Handlers() []any {
	return []any{
		func(c *gin.Context) {
			// response
			c.JSON(200, r.conf) // todo: implement JSON method
		},
		func(c zephyrix.Context) {
			// response
			c.JSON(200, r.conf) // todo: implement JSON method
		},
	}
}

type User struct {
	ID uint64 `orm:"table=users;redisCache;localCache"`
}

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	app := zephyrix.NewApplication()

	app.Database().RegisterEntity(&User{})
	app.RegisterRouteHandler(newRouteHandler)


	app.Router().Group(func(router zephyrix.Router) {
		router.GET("/", func(c *gin.Context) {
			// response
			c.JSON(200, "Hello, World!") // todo: implement JSON method
		})
		router.Group(func(router zephyrix.Router) {
			router.GET("/", func(c *gin.Context) {
				// response
				c.JSON(200, "Hello, World! from /test") // todo: implement JSON method
			})
			router.GET("/kos", func(c *gin.Context) {
				// response
				c.JSON(200, "Hello, World! from /test") // todo: implement JSON method
			})
		}, "/test")
	})

	if err := app.Start(context.Background()); err != nil {
		panic(err)
	}
}
