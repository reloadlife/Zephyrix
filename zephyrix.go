// Project Zephyrix is a Simple yet complex and feature-rich web application framework for Go.
//
// It is designed to be simple to use and easy to understand, but also to be powerful and flexible.
//
// Getting started needs no more than just a few lines of code, but it can be extended to be as complex as you need.
//
// get started by simply installing Zephyrix:
// go install github.com/reloadlife/zephyrix
// and then create a new project:
// zephyrix new <project name>
//
// this command will generate a new project with all the scaffolding you need to get started.
// you can start working on your project rightaway.
//
// in order to run the project in deveopment mode, you can run:
// make dev
// this will start the server in development mode, and you can start working on your project.
// the dev server uses `air` to watch for changes and restart the server.
//
// to start working on your project, you can start by changing the contents of the `main.go` file.
//
// have fun coding :)
package zephyrix

import (
	"context"
	"time"

	"github.com/latolukasz/beeorm/v3"
)

// Zephyrix is the interface that users will see outside of the package
// it might not be wise, to expose an interface, but will see in test phases
type Zephyrix interface {
	// Start the Zephyrix server
	// This method is blocking
	// It will return an error if the server fails to start
	// It will return error, if the server crashes
	// It will return nil if the server is stopped, gracefully, by calling Stop, or by a signal
	Start(context.Context) error

	// Stop the Zephyrix server
	// This method is blocking
	// It will return an error if the server fails to stop
	// It will return nil if the server is stopped, and cleaned up successfully
	Stop() error

	// Cleanup will stop the server,
	// and clean up any resources
	// Flush any data that is in memory to the disk or to the database (if any)
	// This method is blocking until the server is stopped
	// It will return an error if the server fails to be stopped
	// or if something goes wrong during cleanup
	// It will return nil if the server is stopped, and cleaned up successfully.
	Cleanup() error

	// Database will return the Database interface
	Database() Database
	RegisterEntity(entity ...interface{})

	// Router will return a Router instance
	Router() Router
	// RegisterRouteHandler will register a route handler, the handler must implement RouteHandler interface
	RegisterRouteHandler(handlers ...any)
	RegisterMiddleware(middlewares ...any)

	RegisterJob(job ...JobInterface)
	RegisterSchedule(schedule ...ScheduleInterface)
	RegisterCronFunc(spec string, f func())

	RegisterExecuteLaterFunc(duration time.Duration, f func())
}

// Database is the interface that will be used to interact with the database
type Database interface {
	RegisterEntity(entity ...interface{})
	GetEngine() beeorm.Engine
}

// Router is the interface that will be used to define routes
type Router interface {
	Group(func(router Router), ...any)
	GET(relativePath string, handlerFunction any, middlewareFunctions ...any)
	POST(relativePath string, handlerFunction any, middlewareFunctions ...any)
	PUT(relativePath string, handlerFunction any, middlewareFunctions ...any)
	DELETE(relativePath string, handlerFunction any, middlewareFunctions ...any)
	PATCH(relativePath string, handlerFunction any, middlewareFunctions ...any)
	OPTIONS(relativePath string, handlerFunction any, middlewareFunctions ...any)
	HEAD(relativePath string, handlerFunction any, middlewareFunctions ...any)
	CONNECT(relativePath string, handlerFunction any, middlewareFunctions ...any)
	TRACE(relativePath string, handlerFunction any, middlewareFunctions ...any)
	Any(relativePath string, handlerFunction any, middlewareFunctions ...any)
	Match(httpMethods []HTTPVerb, relativePath string, handlerFunction any, middlewareFunctions ...any)
}

// Context is the interface that will be used to interact with the request and response
type Context interface {
	JSON(code int, obj interface{})
}
