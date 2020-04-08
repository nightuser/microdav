package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"golang.org/x/net/context"
)

var (
	logger      = log.New(os.Stdout, "Info:  ", log.LstdFlags)
	errorLogger = log.New(os.Stdout, "Error: ", log.LstdFlags)
)

var db *sql.DB

func showIndex(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hey!"))
}

func startServer(shutdownCompleted *sync.WaitGroup) *http.Server {
	r := mux.NewRouter()
	r.HandleFunc("/", basicAuth(showIndex, "Protected"))
	loggedRouter := handlers.LoggingHandler(os.Stdout, r)

	srv := &http.Server{
		Addr:    ":8000",
		Handler: loggedRouter,
	}

	go func() {
		defer shutdownCompleted.Done()

		logger.Println("Starting server")
		if err := srv.ListenAndServe(); err != nil {
			if err == http.ErrServerClosed {
				logger.Println("Server shut down")
				return
			}
			errorLogger.Fatal(err)
		}
	}()

	return srv
}

func main() {
	var err error
	db, err = sql.Open("sqlite3", "test.db")
	if err != nil {
		errorLogger.Fatal(err)
	}
	defer db.Close()

	shutdownCompleted := &sync.WaitGroup{}
	shutdownCompleted.Add(1)

	srv := startServer(shutdownCompleted)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		errorLogger.Fatal(err)
	}
	shutdownCompleted.Wait()
}
