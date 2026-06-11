package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ChinthaVamsidharReddy/ainyx-backend-task/config"
	dbsqlc "github.com/ChinthaVamsidharReddy/ainyx-backend-task/db/sqlc"
	"github.com/ChinthaVamsidharReddy/ainyx-backend-task/internal/handler"
	"github.com/ChinthaVamsidharReddy/ainyx-backend-task/internal/logger"
	"github.com/ChinthaVamsidharReddy/ainyx-backend-task/internal/repository"
	"github.com/ChinthaVamsidharReddy/ainyx-backend-task/internal/routes"
	"github.com/ChinthaVamsidharReddy/ainyx-backend-task/internal/service"
	"github.com/gofiber/fiber/v2"
	_ "github.com/lib/pq" // PostgreSQL driver – blank import registers it with database/sql
	"go.uber.org/zap"
)

func main() {
	// ── 1. Config ──────────────────────────────────────────────────────────────
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	// ── 2. Logger (Uber Zap) ───────────────────────────────────────────────────
	zapLogger, err := logger.New()
	if err != nil {
		log.Fatalf("logger: %v", err)
	}
	defer zapLogger.Sync() //nolint:errcheck

	// ── 3. Database connection ─────────────────────────────────────────────────
	sqlDB, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		zapLogger.Fatal("failed to open database connection", zap.Error(err))
	}
	defer sqlDB.Close()

	if err := sqlDB.Ping(); err != nil {
		zapLogger.Fatal("database unreachable – check your env vars", zap.Error(err))
	}
	zapLogger.Info("connected to postgres", zap.String("host", cfg.DBHost), zap.String("db", cfg.DBName))

	// ── 4. Wire layers ─────────────────────────────────────────────────────────
	queries := dbsqlc.New(sqlDB)
	repo := repository.New(queries)
	svc := service.New(repo, zapLogger)
	h := handler.New(svc, zapLogger)

	// ── 5. Fiber app ───────────────────────────────────────────────────────────
	app := fiber.New(fiber.Config{
		// Return a plain JSON body on panics instead of crashing.
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			var e *fiber.Error
			if ok := errors.As(err, &e); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{"error": err.Error()})
		},
	})

	routes.Register(app, h, zapLogger)

	// ── 6. Graceful shutdown ───────────────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		addr := fmt.Sprintf(":%s", cfg.ServerPort)
		zapLogger.Info("server starting", zap.String("addr", addr))
		if err := app.Listen(addr); err != nil {
			zapLogger.Fatal("server error", zap.Error(err))
		}
	}()

	<-quit
	zapLogger.Info("shutting down gracefully…")
	if err := app.Shutdown(); err != nil {
		zapLogger.Error("shutdown error", zap.Error(err))
	}
}
