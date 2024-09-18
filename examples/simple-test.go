package main

import (
	"context"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/gin-gonic/gin"
	"go.mamad.dev/zephyrix"
)

// THIS IS NOT SAFE, NEVER DO IN PROD, THIS IS ONLY HERE TO TEST IF DI WORKS
type routeHandler struct {
	conf *zephyrix.Config

	name string
}

func newRouteHandler(conf *zephyrix.Config) *routeHandler {
	return &routeHandler{
		conf: conf,
		name: "hello_world",
	}
}

func (r *routeHandler) Method() []string {
	return []string{"GET"}
}

func (r *routeHandler) Name() string {
	return r.name
}

func (r *routeHandler) Path() string {
	return "/hello_world_di"
}

func (r *routeHandler) Handlers() []any {
	return []any{
		"mw:1,2,3",
		func(c *gin.Context) { // also works with zephyrix.Context
			// response
			c.JSON(200, r.conf) // todo: implement JSON method
		},
	}
}

type mwHandler struct {
}

func newMwHandler() *mwHandler {
	return &mwHandler{}
}

func (m *mwHandler) Name() string {
	return "mw"
}

func (m *mwHandler) Handler(args ...any) any {
	return func(c *gin.Context) {
		spew.Dump(args)
		println("middleware 1")
		c.Next()
	}
}

func newRouteHandler2(conf *zephyrix.Config) *routeHandler {
	return &routeHandler{
		conf: conf,
		name: "testing_hw",
	}
}

type User struct {
	ID uint64 `orm:"table=users;redisCache;localCache"`
}

func main() {
	app := zephyrix.NewApplication()

	app.Database().RegisterEntity(&User{})
	app.RegisterRouteHandler(newRouteHandler)
	app.RegisterRouteHandler(newRouteHandler2)
	app.RegisterMiddleware(newMwHandler)

	app.Router().Group(func(router zephyrix.Router) {
		router.GET("/", func(c *gin.Context) {
			c.JSON(200, "Hello, World!")
		})
		router.Group(func(router zephyrix.Router) {
			router.GET("/", func(z zephyrix.Context) {
				z.JSON(200, "Hello, World! from /test")
			})
			router.GET("/second-endpoint", func(z zephyrix.Context) {
				z.JSON(200, "Hello, World! from /test/second-endpoint")
			})
		}, "/test")
	})

	app.RegisterCronFunc("* * * * * *", func() {
		zephyrix.Logger.Debug("CronJob Triggered")
	})

	app.RegisterExecuteLaterFunc(5*time.Second, func() {
		zephyrix.Logger.Debug("ExecuteLaterFunc Triggered")
	})

	if err := app.Start(context.Background()); err != nil {
		panic(err)
	}
}
