package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type application struct {
	ErrorLog *log.Logger
}

func main() {
	fmt.Println("Calling API..")

	Newlogger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	app := &application{
		ErrorLog: Newlogger,
	}

	srv := &http.Server{
		Addr:        ":8080",
		Handler:     app.routes(),
		IdleTimeout: time.Minute,
	}

	err := srv.ListenAndServe()
	Newlogger.Fatal(err)
}
