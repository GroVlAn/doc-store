package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/GroVlAn/doc-store/internal/config"
	"github.com/GroVlAn/doc-store/internal/handler"
	mongoclient "github.com/GroVlAn/doc-store/internal/mongo"
	repository "github.com/GroVlAn/doc-store/internal/repostiory"
	"github.com/GroVlAn/doc-store/internal/server"
	"github.com/GroVlAn/doc-store/internal/service"
	"github.com/rs/zerolog"
)

const (
	defaultConfigPath = "configs/config-dev.yaml"
)

func main() {
	timeStart := time.Now()

	configPath := flag.String("config", defaultConfigPath, "Path to the configuration file")
	flag.Parse()

	l := zerolog.New(os.Stdout).
		With().
		Timestamp().
		Logger().
		Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "2006-01-02 15:04:05"})

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	err := config.LoadEnv()
	if err != nil {
		l.Fatal().Err(err).Msg("failed to load env")
	}

	cfg, err := config.New(*configPath)
	if err != nil {
		l.Fatal().Err(err).Msg("failed to load configuration")
	}

	mgc := mongoclient.New(cfg.Mongo.URI)

	err = mgc.Connect(ctx)
	if err != nil {
		l.Fatal().Err(err).Msg("failed to connect to mongo db")
	}

	r := repository.New(mgc.Client())

	fr := repository.NewFileRepository()

	s := service.New(service.Deps{
		UserRepo:       r,
		TokenRepo:      r,
		DocumentRepo:   r,
		FileRepo:       fr,
		DefaultTimeout: cfg.Service.DefaultTimeout,
		HashCost:       cfg.Service.HashCost,
		TokenEndTTL:    cfg.Service.TokenEndTTl,
		SecretKey:      cfg.Service.SecretKey,
	})

	h := handler.New(l, s, s)

	server := server.New(
		h.Handler(),
		server.Deps{
			Port:              cfg.HTTP.Port,
			MaxHeaderBytes:    cfg.HTTP.MaxHeaderBytes,
			ReadHeaderTimeout: time.Duration(cfg.HTTP.ReadHeaderTimeout) * time.Second,
			WriteTimeout:      time.Duration(cfg.HTTP.WriteTimeout) * time.Second,
		},
	)

	go func() {
		if err := server.Strart(); err != nil && err != http.ErrServerClosed {
			l.Fatal().Err(err).Msg("failed to start server")
		}
	}()

	l.Info().Msgf("server start on port: %s load time: %v", cfg.HTTP.Port, time.Since(timeStart))

	<-ctx.Done()
	l.Info().Msg("shutting down")
}
