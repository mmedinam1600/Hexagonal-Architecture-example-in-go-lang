package main

import (
	"context"
	"net/http"
	"os"
	"time"

	inhttp "hexagonal-bank/internal/adapters/in/http"
	"hexagonal-bank/internal/adapters/out/eventbus"
	"hexagonal-bank/internal/adapters/out/memory"
	"hexagonal-bank/internal/adapters/out/stp"
	"hexagonal-bank/internal/platform/logging"
)

func main() {
	// Logger
	applicationLogger := logging.NewStd()

	// In-memory repository simulating a database
	accountRepository := memory.NewAccountRepo()

	// Local event bus (in-process) simulating a queue/broker
	localEventBus := eventbus.NewLocalBus(applicationLogger)

	// Fake STP client with retry/backoff + jitter
	fakeSTPGateway := stp.NewFakeSTP(applicationLogger)

	// HTTP API wiring: inject implementations into ports
	httpAPI := inhttp.NewAPI(applicationLogger, accountRepository, accountRepository, fakeSTPGateway, localEventBus)

	httpServer := &http.Server{
		Addr:              ":8080",
		Handler:           httpAPI.Router(),
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	applicationLogger.Info("HexBank API starting on :8080...")
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		applicationLogger.Error("server error", "err", err)
		os.Exit(1)
	}

	// Graceful shutdown example (not used in this minimal main)
	_ = context.Background()
}
