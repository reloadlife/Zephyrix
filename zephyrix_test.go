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

	redisC, err := setupRedis()
	if err != nil {
		log.Fatalf("Failed to start Redis: %s", err)
	}
	mysqlC, err := setupMySQL()
	if err != nil {
		log.Fatalf("Failed to start MySQL: %s", err)
	}

	// Get the connection details for each service
	redisHost, _ := redisC.Host(ctx)
	redisPort, _ := redisC.MappedPort(ctx, "6379")
	mysqlHost, _ := mysqlC.Host(ctx)
	mysqlPort, _ := mysqlC.MappedPort(ctx, "3306")

	// Set the configurations for the testApp
	zephyrix.TestConfig = &zephyrix.Config{
		Database: zephyrix.DatabaseConfig{
			Pools: []zephyrix.DatabasePoolConfig{
				{
					Name: "default",
					DSN:  "root:password@tcp(" + mysqlHost + ":" + mysqlPort.Port() + ")/testdb",
					Cache: zephyrix.CacheConfig{
						Enabled: true,
						Size:    100,
					},
					Redis: zephyrix.RedisConfig{
						Enabled:  true,
						Address:  redisHost + ":" + redisPort.Port(),
						Username: "",
						Password: "",
						DB:       0,
					},
				},
			},
		},
	}

	// Initialize testApp with the configurations
	testApp = zephyrix.NewApplication()

	// Run the tests
	exitCode := m.Run()

	// Clean up
	if err := testApp.Stop(); err != nil {
		log.Printf("Error during cleanup: %s", err)
	}

	err = redisC.Terminate(ctx)
	if err != nil {
		log.Printf("Error during cleanup (failed to kill redis): %s (%s)", err, redisC.GetContainerID())
	}
	err = mysqlC.Terminate(ctx)
	if err != nil {
		log.Printf("Error during cleanup (failed to kill mysql): %s (%s)", err, mysqlC.GetContainerID())
	}
	// Exit with the status code from the test run
	os.Exit(exitCode)
}
