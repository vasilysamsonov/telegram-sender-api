package app

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"telegram-sender-api/config"
	httppkg "telegram-sender-api/internal/controller/http"
	telegram "telegram-sender-api/internal/repo/webapi/telegram"
	"telegram-sender-api/internal/usecase/message"
	"telegram-sender-api/pkg/logger"

	"github.com/gofiber/fiber/v2"
)

func Run(cfg *config.Config) error {
	l := logger.New(cfg.Log.Level)

	httpClient := &http.Client{
		Timeout: cfg.HTTP.Timeout,
	}

	messageUseCase := message.New(telegram.New(httpClient))

	server := fiber.New()
	httppkg.NewRouter(server, messageUseCase, l)

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Listen(":" + cfg.HTTP.Port)
	}()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case <-interrupt:
		l.Info("shutdown signal received")
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("listen server: %w", err)
		}
		return nil
	}

	if err := server.ShutdownWithTimeout(cfg.HTTP.ShutdownTimeout); err != nil {
		return fmt.Errorf("shutdown server: %w", err)
	}

	return nil
}
