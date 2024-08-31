package main

import (
	"context"
	"log"

	"net/http"
	_ "net/http/pprof"

	"go.mamad.dev/zephyrix"
)

type User struct {
	ID uint64 `orm:"table=users;redisCache;localCache"`
}

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	app := zephyrix.NewApplication()

	app.Database().RegisterEntity(&User{})

	if err := app.Start(context.Background()); err != nil {
		panic(err)
	}

}
