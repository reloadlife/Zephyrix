package main

import (
	"context"
	"log"

	"net/http"
	_ "net/http/pprof"

	"go.mamad.dev/zephyrix"
)

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	app := zephyrix.NewApplication()

	if err := app.Start(context.Background()); err != nil {
		panic(err)
	}

}
