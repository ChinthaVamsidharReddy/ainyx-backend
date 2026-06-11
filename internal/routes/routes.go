package routes

import (
	"github.com/ChinthaVamsidharReddy/ainyx-backend-task/internal/handler"
	"github.com/ChinthaVamsidharReddy/ainyx-backend-task/internal/middleware"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// Register attaches all application routes to the Fiber app.
func Register(app *fiber.App, h *handler.UserHandler, logger *zap.Logger) {
	// Global middleware applied before every handler.
	app.Use(middleware.RequestID())
	app.Use(middleware.Logger(logger))

	// Health-check – useful for Docker / k8s probes.
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// User resource routes.
	users := app.Group("/users")
	users.Post("/", h.CreateUser)
	users.Get("/", h.ListUsers)
	users.Get("/:id", h.GetUser)
	users.Put("/:id", h.UpdateUser)
	users.Delete("/:id", h.DeleteUser)
}
