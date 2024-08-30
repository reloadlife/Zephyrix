package zephyrix

import (
	"log"
	"os"
)

// Logger is the logger that will be used by the Zephyrix server
// the current logger is Temporary, and will be replaced by a better logger
// logger will be able to log to
// - a file,
// - a database,
// - a remote server
// ? or any other place that the user wants to log to
var Logger = log.New(os.Stdout, "Zephyrix: ", log.LstdFlags)
