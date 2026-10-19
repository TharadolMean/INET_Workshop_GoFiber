package controller

import (
	"errors"
	"log"
	"strconv"

	"gofiber-workshop/database"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func apiError(ctx *fiber.Ctx, status int, message string) error {
	return ctx.Status(status).JSON(fiber.Map{"success": false, "message": message})
}

// ใช้กับ Route ใหม่เท่านั้น เพื่อให้ข้อ 5–6 ยังทำงานได้เมื่อ MySQL ไม่พร้อม
func RequireDatabase(ctx *fiber.Ctx) error {
	if database.DBConn == nil {
		return apiError(ctx, fiber.StatusInternalServerError, "Database unavailable; check server configuration")
	}
	return ctx.Next()
}

func databaseError(ctx *fiber.Ctx, err error) error {
	log.Printf("Database operation failed: %v", err)
	return apiError(ctx, fiber.StatusInternalServerError, "Database operation failed")
}

func recordError(ctx *fiber.Ctx, err error, name string) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return apiError(ctx, fiber.StatusNotFound, name+" not found")
	}
	return databaseError(ctx, err)
}

func parseID(text string) (uint, error) {
	id, err := strconv.ParseUint(text, 10, strconv.IntSize)
	if err != nil || id == 0 {
		return 0, errors.New("id must be a positive integer")
	}
	return uint(id), nil
}
