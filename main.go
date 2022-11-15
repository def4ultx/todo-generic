package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {

	connection := "postgresql://postgres:postgres@localhost/postgres?sslmode=disable"
	config, err := pgxpool.ParseConfig(connection)
	if err != nil {
		log.Fatal("error configuring the database: ", err)
	}
	conn, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Fatal("error connecting to the database: ", err)
	}
	defer conn.Close()

	err = conn.Ping(context.Background())
	if err != nil {
		log.Fatal("cannot ping to db", err)
	}

	r := mux.NewRouter()
	r.Use(DatabaseMiddleware(conn), ValidatorMiddleware())
	r.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`OK`))
	})

	r.Path("/todos").Methods(http.MethodGet).Handler(Handle(ListTodo))
	r.Path("/todos").Methods(http.MethodPost).Handler(Handle(CreateTodo))
	r.Path("/todos/{id}").Methods(http.MethodGet).Handler(Handle(GetTodo))
	r.Path("/todos/{id}").Methods(http.MethodPut).Handler(Handle(UpdateTodo))
	r.Path("/todos/{id}").Methods(http.MethodDelete).Handler(Handle(DeleteTodo))

	srv := &http.Server{
		Addr:         "0.0.0.0:8080",
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r,
	}

	go func() {
		log.Println("The service is ready to listen and serve.")
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	wait := 15 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	err = srv.Shutdown(ctx)
	if err != nil {
		log.Fatal("error occur while shutting down", err)
	}

	log.Println("shutting down")
	os.Exit(0)
}
