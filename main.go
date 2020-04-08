package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/nightuser/dav-server/usermanager"
	"golang.org/x/net/context"
	"golang.org/x/net/webdav"
)

var (
	logger      = log.New(os.Stdout, "Info:  ", log.LstdFlags)
	errorLogger = log.New(os.Stdout, "Error: ", log.LstdFlags)
)

var userManager *usermanager.UserManager

func showIndex(w http.ResponseWriter, r *http.Request) {
	username, password, _ := r.BasicAuth()
	w.Write([]byte(fmt.Sprintf("Hey %s : %s!", username, password)))
}

func startServer(shutdownCompleted *sync.WaitGroup) *http.Server {
	dir := "./foo"
	webdavHandler := &webdav.Handler{
		Prefix:     "/dav/",
		FileSystem: webdav.Dir(dir),
		LockSystem: webdav.NewMemLS(),
		Logger: func(r *http.Request, err error) {
			if err == nil {
				logger.Printf("webdav %s\n", r.URL)
			} else {
				errorLogger.Printf(
					"webdav %s %s\n", r.URL, err)
			}
		},
	}

	realm := "Protected"
	davHandler := accessHandler(webdavHandler.ServeHTTP, realm)

	r := mux.NewRouter()
	r.HandleFunc("/", showIndex)
	r.PathPrefix("/dav/{username}/").HandlerFunc(davHandler)
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
	userManager = usermanager.New("test.db")
	defer userManager.Close()

	shutdownCompleted := &sync.WaitGroup{}
	shutdownCompleted.Add(1)

	srv := startServer(shutdownCompleted)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	checkError(srv.Shutdown(ctx))
	shutdownCompleted.Wait()
}
