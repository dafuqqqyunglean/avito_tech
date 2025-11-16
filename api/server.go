package api

import (
	"context"
	"net/http"
	"time"

	"github.com/dafuqqqyunglean/avito_tech/api/handler"
	"github.com/dafuqqqyunglean/avito_tech/api/middleware"
	"github.com/dafuqqqyunglean/avito_tech/config"
	prserv "github.com/dafuqqqyunglean/avito_tech/service/pr"
	teamserv "github.com/dafuqqqyunglean/avito_tech/service/team"
	userserv "github.com/dafuqqqyunglean/avito_tech/service/user"
	"github.com/gorilla/mux"
)

const (
	maxHeaderBytes = 1 << 20
	readTimeout    = 10 * time.Second
	writeTimeout   = 10 * time.Second
)

type Server struct {
	httpServer *http.Server
	router     *mux.Router
}

func NewServer(ctx context.Context, config config.Config) *Server {
	router := mux.NewRouter()

	wrappedRouter := middleware.RecoveryMiddleware(ctx, router)

	return &Server{
		httpServer: &http.Server{
			Addr:           config.ServerPort,
			MaxHeaderBytes: maxHeaderBytes,
			ReadTimeout:    readTimeout,
			WriteTimeout:   writeTimeout,
			Handler:        wrappedRouter,
		},
		router: router,
	}
}

func (s *Server) Run() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) HandleRoutes(ctx context.Context, teamService teamserv.Service, userService userserv.Service, prService prserv.Service) {
	s.router.HandleFunc("/team/add", handler.CreateTeam(ctx, teamService)).Methods(http.MethodPost)
	s.router.HandleFunc("/team/get", handler.GetTeam(ctx, teamService)).Methods(http.MethodGet)
	s.router.HandleFunc("/users/setIsActive", handler.SetActive(ctx, userService)).Methods(http.MethodPost)
	s.router.HandleFunc("/users/getReview", handler.GetReview(ctx, userService)).Methods(http.MethodGet)
	s.router.HandleFunc("/pullRequest/create", handler.CreatePullRequest(ctx, prService)).Methods(http.MethodPost)
	s.router.HandleFunc("/pullRequest/merge", handler.SetMerged(ctx, prService)).Methods(http.MethodPost)
	s.router.HandleFunc("/pullRequest/reassign", handler.Reassign(ctx, prService)).Methods(http.MethodPost)
}
