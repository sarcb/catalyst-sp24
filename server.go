package catalyst

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	maut "github.com/jonas-plum/maut/auth"

	"github.com/sarcb/catalyst-sp24/bus"
	"github.com/sarcb/catalyst-sp24/busservice"
	"github.com/sarcb/catalyst-sp24/database"
	"github.com/sarcb/catalyst-sp24/generated/api"
	"github.com/sarcb/catalyst-sp24/hooks"
	"github.com/sarcb/catalyst-sp24/index"
	"github.com/sarcb/catalyst-sp24/service"
	"github.com/sarcb/catalyst-sp24/storage"
)

type Config struct {
	IndexPath string
	DB        *database.Config
	Storage   *storage.Config

	Auth            *maut.Config
	ExternalAddress string
	InternalAddress string
	Network         string
	Port            int
}

type Server struct {
	Bus     *bus.Bus
	DB      *database.Database
	Index   *index.Index
	Storage *storage.Storage
	Server  chi.Router
}

func New(hooks *hooks.Hooks, config *Config) (*Server, error) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Minute*10)
	defer cancel()

	catalystStorage, err := storage.New(config.Storage)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage: %w", err)
	}

	catalystIndex, err := index.New(config.IndexPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create index: %w", err)
	}

	catalystBus := bus.New()

	catalystDatabase, err := database.New(ctx, catalystIndex, catalystBus, hooks, config.DB)
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}

	busservice.New(config.InternalAddress+"/api", config.Auth.InitialAPIKey, config.Network, catalystBus, catalystDatabase)

	catalystService, err := service.New(catalystBus, catalystDatabase, catalystStorage, GetVersion())
	if err != nil {
		return nil, fmt.Errorf("failed to create service: %w", err)
	}

	authenticator, err := maut.NewAuthenticator(ctx, config.Auth, newCatalystResolver(catalystDatabase))
	if err != nil {
		return nil, fmt.Errorf("failed to create authenticator: %w", err)
	}

	apiServer, err := setupAPI(authenticator, catalystService, catalystStorage, catalystDatabase, catalystBus, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create api server: %w", err)
	}

	return &Server{
		Bus:     catalystBus,
		DB:      catalystDatabase,
		Index:   catalystIndex,
		Storage: catalystStorage,
		Server:  apiServer,
	}, nil
}

func setupAPI(authenticator *maut.Authenticator, catalystService *service.Service, catalystStorage *storage.Storage, catalystDatabase *database.Database, bus *bus.Bus, config *Config) (chi.Router, error) {
	middlewares := []func(next http.Handler) http.Handler{
		authenticator.Authenticate(),
		authenticator.AuthorizeBlockedUser(),
	}

	// create server
	apiServer := api.NewServer(catalystService, permissionAuth(authenticator), middlewares...)
	apiServer.Mount("/files", fileServer(authenticator, catalystDatabase, bus, catalystStorage, config))

	server := chi.NewRouter()
	server.Use(middleware.RequestID, middleware.RealIP, middleware.Logger, middleware.Recoverer)
	server.Mount("/api", apiServer)
	server.Mount("/auth", authenticator.Server())
	server.With(middlewares...).Handle("/wss", handleWebSocket(bus))

	server.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/ui/", http.StatusFound)
	})

	return server, nil
}

func permissionAuth(authenticator *maut.Authenticator) func([]string) func(http.Handler) http.Handler {
	return func(strings []string) func(http.Handler) http.Handler {
		return authenticator.AuthorizePermission(strings...)
	}
}

func fileServer(authenticator *maut.Authenticator, catalystDatabase *database.Database, bus *bus.Bus, catalystStorage *storage.Storage, config *Config) *chi.Mux {
	fileRW := authenticator.AuthorizePermission("file:read", "file:write") // TODO: add test
	tudHandler := tusdUpload(catalystDatabase, bus, catalystStorage.S3(), config.ExternalAddress)
	server := chi.NewRouter()
	server.With(fileRW).Head("/{ticketID}/tusd/{id}", tudHandler)
	server.With(fileRW).Patch("/{ticketID}/tusd/{id}", tudHandler)
	server.With(fileRW).Post("/{ticketID}/tusd", tudHandler)
	server.With(fileRW).Post("/{ticketID}/upload", upload(catalystDatabase, catalystStorage.S3(), catalystStorage.Uploader()))
	server.With(fileRW).Get("/{ticketID}/download/{key}", download(catalystStorage.Downloader()))

	return server
}
