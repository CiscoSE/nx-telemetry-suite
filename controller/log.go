package controller

import (
	"log"
	"os"
)

// CustomLog is the method use for manage all logs generated in the system
func CustomLog(message string, severity string) {
	// If debug is turn on, print everything
	if os.Getenv("DEBUG") == "on" {
		log.Print(severity + ": " + message)
	} else {
		// Only print ERROR severity
		if severity == "ERROR" {
			log.Print(severity + ": " + message)
		}
	}
}
