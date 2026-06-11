package handler

import (
	"errors"
	"strconv"

	"github.com/ChinthaVamsidharReddy/ainyx-backend-task/internal/models"
	"github.com/ChinthaVamsidharReddy/ainyx-backend-task/internal/service"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// UserHandler holds the service and validator needed by all user endpoints.
type UserHandler struct {
	svc      *service.UserService
	validate *validator.Validate
	logger   *zap.Logger
}

// New wires a UserHandler.
func New(svc *service.UserService, logger *zap.Logger) *UserHandler {
	return &UserHandler{
		svc:      svc,
		validate: validator.New(),
		logger:   logger,
	}
}

// CreateUser handles POST /users
func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	var req models.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{Error: "invalid JSON body"})
	}

	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(models.ErrorResponse{Error: err.Error()})
	}

	resp, err := h.svc.CreateUser(c.Context(), &req)
	if err != nil {
		h.logger.Error("CreateUser handler error", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{Error: "internal server error"})
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// GetUser handles GET /users/:id
func (h *UserHandler) GetUser(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{Error: "id must be a positive integer"})
	}

	resp, err := h.svc.GetUser(c.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{Error: "user not found"})
		}
		h.logger.Error("GetUser handler error", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{Error: "internal server error"})
	}

	return c.JSON(resp)
}

// UpdateUser handles PUT /users/:id
func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{Error: "id must be a positive integer"})
	}

	var req models.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{Error: "invalid JSON body"})
	}

	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(models.ErrorResponse{Error: err.Error()})
	}

	resp, err := h.svc.UpdateUser(c.Context(), id, &req)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{Error: "user not found"})
		}
		h.logger.Error("UpdateUser handler error", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{Error: "internal server error"})
	}

	return c.JSON(resp)
}

// DeleteUser handles DELETE /users/:id
func (h *UserHandler) DeleteUser(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{Error: "id must be a positive integer"})
	}

	if err := h.svc.DeleteUser(c.Context(), id); err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{Error: "user not found"})
		}
		h.logger.Error("DeleteUser handler error", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{Error: "internal server error"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ListUsers handles GET /users  (supports ?page=1&page_size=10)
func (h *UserHandler) ListUsers(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "10"))

	resp, err := h.svc.ListUsers(c.Context(), page, pageSize)
	if err != nil {
		h.logger.Error("ListUsers handler error", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{Error: "internal server error"})
	}

	// Return bare array when no pagination query params supplied (spec compliance).
	if c.Query("page") == "" && c.Query("page_size") == "" {
		return c.JSON(resp.Data)
	}
	return c.JSON(resp)
}

// parseID extracts and validates the :id URL parameter.
func parseID(c *fiber.Ctx) (int32, error) {
	raw := c.Params("id")
	n, err := strconv.ParseInt(raw, 10, 32)
	if err != nil || n < 1 {
		return 0, errors.New("invalid id")
	}
	return int32(n), nil
}
