package app

import (
	ucrprocessor "barter-port/internal/auth/application/uc-result-inbox-processor"
	authconsumer "barter-port/internal/auth/infrastructure/kafka/consumer"
	authkafka "barter-port/internal/auth/infrastructure/kafka/producer"
	"barter-port/pkg/bootstrap"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/oklog/run"
)

type App struct {
	logger            *slog.Logger
	db                *pgxpool.Pool
	server            *http.Server
	ucResultConsumer  *authconsumer.UCResultInboxConsumer
	ucResultProcessor *ucrprocessor.Processor
	outboxPublisher   *authkafka.UCOutbox
}

func NewApp(cfg bootstrap.Config) (*App, error) {
	app := &App{}
	var err error
	defer func() {
		if err != nil {
			app.Close()
		}
	}()

	err = app.initDatabase(cfg)
	if err != nil {
		return nil, err
	}

	err = app.initServices(cfg)
	if err != nil {
		return nil, err
	}

	return app, nil
}

func (a *App) Run() error {
	var g run.Group
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer a.Close()

	g.Add(func() error {
		return a.outboxPublisher.Run(ctx)
	}, func(error) {
		cancel()
	})

	g.Add(func() error {
		return a.ucResultConsumer.Run(ctx)
	}, func(error) {
		cancel()
	})

	g.Add(func() error {
		return a.ucResultProcessor.Run(ctx)
	}, func(error) {
		cancel()
	})

	g.Add(func() error {
		a.logger.Info("backend listening", slog.String("addr", a.server.Addr))
		return a.server.ListenAndServe()
	}, func(error) {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()

		if err := a.server.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.logger.Error("failed to shutdown auth server", slog.Any("error", err))
		}
	})

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	g.Add(func() error {
		select {
		case <-stop:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}, func(error) {
		signal.Stop(stop)
	})

	err := g.Run()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func (a *App) Close() {
	if a == nil {
		return
	}
	if a.outboxPublisher != nil {
		if err := a.outboxPublisher.Close(); err != nil && a.logger != nil {
			a.logger.Error("failed to close kafka writer", slog.Any("error", err))
		}
	}
	if a.db != nil {
		a.db.Close()
	}
}
