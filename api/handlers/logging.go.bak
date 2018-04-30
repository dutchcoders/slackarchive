package handlers

import (
	"bufio"
	"log"
	"net"
	"net/http"
	"time"
)

type statusLoggingResponseWriter struct {
	status int
	http.ResponseWriter
}

func (w *statusLoggingResponseWriter) Hijack() (rwc net.Conn, buf *bufio.ReadWriter, err error) {
	return w.ResponseWriter.(http.Hijacker).Hijack()
}

func (w *statusLoggingResponseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func LoggingHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		began := time.Now()
		slrw := &statusLoggingResponseWriter{-1, w}
		h.ServeHTTP(slrw, req)
		log.Printf("%s\t%s\t%s\t%s\t%d", req.RemoteAddr, req.Method, req.URL, time.Since(began), slrw.status)
	})
}
