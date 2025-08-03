package utils

import (
	"log"
	"os"
)

var (
	WarningLog *log.Logger
	ErrorLog   *log.Logger
	InfoLog    *log.Logger
)

func init() {
	file, err := os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	WarningLog = log.New(file, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLog = log.New(file, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	InfoLog = log.New(file, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
}
