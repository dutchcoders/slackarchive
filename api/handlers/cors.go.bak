package handlers

import "net/http"

// FIXME proper middleware solution
func CorsHandler(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "content-type, accept, authorization, X-Alt-Referer")
		w.Header().Set("Access-Control-Allow-Methods", "GET,PUT,POST,PATCH,DELETE")
		if r.Method == "OPTIONS" {
			return
		}
		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
