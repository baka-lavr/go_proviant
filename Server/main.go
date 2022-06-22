package main

import (
	"os"
	"os/signal"
	"log"
	"net/http"
	"database/sql"
	"context"
	"time"
)

type Application struct {
	Server *http.Server
	db *sql.DB
	Mail *MailClient
}

func main() {
	var err error
	app := new(Application)
	router := app.Route()
	server := &http.Server {
		Addr: ":4000",
		Handler: router,
	}
	app.Server = server
	app.db, err = OpenDB()
	if err != nil {
		log.Fatal(err)
	}
	app.Mail, err = app.InitMail()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Routing")

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt)
		<- sig
		ctx, cancel := context.WithTimeout(context.Background(),time.Second*10)
		defer cancel()
		app.db.Close()
		app.Mail.done<-1
		app.Server.Shutdown(ctx)
		log.Println("Server stopped")
	}()
	
	err = app.Server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}