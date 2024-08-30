package zephyrix_test

import (
	"os"
	"testing"

	"context"
	"log"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.mamad.dev/zephyrix"
)

// well well well, every project will need tests !
// we will be using https://testcontainers.com/ to run
// a REDIS, and a MySQL, MongoDB, and any other nessery services
// to test the Zephyrix server.
// we will also be using it to run a Zephyrix server to test **Some** edge cases
// to make sure everything is working as expected.

var testApp zephyrix.Zephyrix
var ctx context.Context

func setupRedis() (testcontainers.Container, error) {
	req := testcontainers.ContainerRequest{
		Image:        "redis:latest",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	}
	return testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
}

func setupMySQL() (testcontainers.Container, error) {
	req := testcontainers.ContainerRequest{
		Image:        "mysql:latest",
		ExposedPorts: []string{"3306/tcp"},
		Env: map[string]string{
			"MYSQL_ROOT_PASSWORD": "password",
			"MYSQL_DATABASE":      "testdb",
		},
		WaitingFor: wait.ForLog("MySQL init process done. Ready for start up."),
	}
	return testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
}

func setupMongoDB() (testcontainers.Container, error) {
	req := testcontainers.ContainerRequest{
		Image:        "mongo:latest",
		ExposedPorts: []string{"27017/tcp"},
		WaitingFor:   wait.ForLog("Waiting for connections"),
	}
	return testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
}

func setupZephyrixServer() (testcontainers.Container, error) {
	req := testcontainers.ContainerRequest{
		Image:        "zephyrix:latest", // Assuming you have a Docker image for Zephyrix
		ExposedPorts: []string{"8080/tcp"},
		WaitingFor:   wait.ForHTTP("/health").WithPort("8080/tcp"),
	}
	return testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
}

// TestMain is the entry point for the tests
// we will be initializing the testApp here
// and we will be running the tests
// before initializing the testApp,
// we need to initialize the test containers,
// and also set them up (as configuration) for the testApp
// and after the tests are done, we need to clean up the test containers
// and also clean up the testApp
// and then exit the tests
func TestMain(m *testing.M) {
	ctx = context.Background()

	// Initialize test containers

	// we wont be using these services for now

	// redisC, err := setupRedis()
	// if err != nil {
	// 	log.Fatalf("Failed to start Redis: %s", err)
	// }
	// defer redisC.Terminate(ctx)

	// mysqlC, err := setupMySQL()
	// if err != nil {
	// 	log.Fatalf("Failed to start MySQL: %s", err)
	// }
	// defer mysqlC.Terminate(ctx)

	// mongoC, err := setupMongoDB()
	// if err != nil {
	// 	log.Fatalf("Failed to start MongoDB: %s", err)
	// }
	// defer mongoC.Terminate(ctx)

	// zephyrixC, err := setupZephyrixServer()
	// if err != nil {
	// 	log.Fatalf("Failed to start Zephyrix server: %s", err)
	// }
	// defer zephyrixC.Terminate(ctx)

	// Get the connection details for each service
	// redisHost, _ := redisC.Host(ctx)
	// redisPort, _ := redisC.MappedPort(ctx, "6379")
	// mysqlHost, _ := mysqlC.Host(ctx)
	// mysqlPort, _ := mysqlC.MappedPort(ctx, "3306")
	// mongoHost, _ := mongoC.Host(ctx)
	// mongoPort, _ := mongoC.MappedPort(ctx, "27017")
	// zephyrixHost, _ := zephyrixC.Host(ctx)
	// zephyrixPort, _ := zephyrixC.MappedPort(ctx, "8080")

	// Initialize testApp with the configurations
	config := zephyrix.Config{}
	testApp = zephyrix.NewApplication(&config)

	// Run the tests
	exitCode := m.Run()

	// Clean up
	if err := testApp.Cleanup(); err != nil {
		log.Printf("Error during cleanup: %s", err)
	}

	// Exit with the status code from the test run
	os.Exit(exitCode)
}

// Add your test functions here
func TestSomething(t *testing.T) {
	// Your test code here
}

func TestEdgeCase(t *testing.T) {
	// Test using the Zephyrix server running in a container
}
