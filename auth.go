package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/nightuser/microdav/usermanager"
)

func accessHandler(handler http.HandlerFunc, realm string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		requestedUsername := vars["username"]

		username, password, ok := r.BasicAuth()
		if !ok {
			unauthorize(w, realm)
			return
		}

		if username != requestedUsername {
			errorLogger.Printf(
				"'%s' tried to accees '%s's files",
				username, requestedUsername)
			unauthorize(w, realm)
			return
		}

		if err := userManager.CheckPassword(username, password); err != nil {
			if err != usermanager.ErrPasswordDoesNotMatch {
				errorLogger.Printf(
					"Wrong password for user '%s'\n",
					username)
			}
			unauthorize(w, realm)
			return
		}

		handler(w, r)
	}
}

func unauthorize(w http.ResponseWriter, realm string) {
	w.Header().Set(
		"WWW-Authenticate", fmt.Sprintf(`Basic realm="%s"`, realm))
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte("Unauthorized"))
}
