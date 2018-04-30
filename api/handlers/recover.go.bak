package handlers

import (
	"fmt"
	"net/http"
	"runtime"
)

func RecoverHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				trace := make([]byte, 1024)
				count := runtime.Stack(trace, true)
				fmt.Printf("Recover from panic: %s\n", err)
				fmt.Printf("Stack of %d bytes: %s\n", count, trace)
				http.Error(w, http.StatusText(500), 500) // What we had before
				return
			}
		}()

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}
