package server

import (
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/willlay0203/gohttp/middleware"

	"github.com/rs/cors"
)

type Server struct {
	Mux         *http.ServeMux
	Port        string
	middlewares []middleware.Middleware
}

type apiFunctionWithError func(w http.ResponseWriter, r *http.Request) error

func Setup(port string) Server {
	m := Server{
		Mux:  http.NewServeMux(),
		Port: ":" + port,
	}

	m.Enable(middleware.RequestId())

	return m
}

func (mux *Server) Start() {
	m := middleware.Adapt(mux.Mux, mux.middlewares...)

	// Wrap mux with CORs middleware
	handler := cors.Default().Handler(m)

	// Sets the default logging behaviour
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	log.Fatal(http.ListenAndServe(mux.Port, handler))
}

func (mux *Server) Enable(a middleware.Middleware) {
	mux.middlewares = append(mux.middlewares, a)
}

func (mux *Server) GET(p string, handler apiFunctionWithError) {
	mux.Mux.HandleFunc("GET "+p, convertToApiFunctionWithError(handler))
}

func (mux *Server) POST(p string, handler apiFunctionWithError) {
	mux.Mux.HandleFunc("POST "+p, convertToApiFunctionWithError(handler))
}

func (mux *Server) PUT(p string, handler apiFunctionWithError) {
	mux.Mux.HandleFunc("PUT "+p, convertToApiFunctionWithError(handler))
}

func (mux *Server) PATCH(p string, handler apiFunctionWithError) {
	mux.Mux.HandleFunc("PATCH "+p, convertToApiFunctionWithError(handler))
}

func (mux *Server) DELETE(p string, handler apiFunctionWithError) {
	mux.Mux.HandleFunc("DELETE "+p, convertToApiFunctionWithError(handler))
}

func convertToApiFunctionWithError(a apiFunctionWithError) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := a(w, r); err != nil {
			if e, ok := err.(APIError); ok {
				// This is the internal	log
				slog.Error(
					"API Response",
					"status", e.StatusCode,
					"error", e,
				)
				// HTTP Response
				http.Error(w, e.Error(), e.StatusCode)
			} else {
				// Default error
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}
	}
}
