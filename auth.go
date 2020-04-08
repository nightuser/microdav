package main

import (
	"fmt"
	"net/http"
)

func basicAuth(handler http.HandlerFunc, realm string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()

		if !ok {
			unauthorize(w, realm)
			return
		}

		if err := checkPassword(username, password); err != nil {
			if err != ErrPasswordDoesNotMatch {
				errorLogger.Print(err)
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
