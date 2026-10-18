package main

import (
	"log"

	// Alias c ทำให้เรียก Handler ใน Package controller ด้วยชื่อสั้น เช่น c.TestParams
	c "gofiber-workshop/controller"

	"github.com/gofiber/fiber/v2"
	fiberAuth "github.com/gofiber/fiber/v2/middleware/basicauth"
)

func main() {
	app := fiber.New()

	// Basic Auth ทำงานก่อนเข้า Handler ทุก Route ที่อยู่ใต้ /api
	// ชื่อผู้ใช้และรหัสผ่านนี้เป็นค่าฝึก Workshop เท่านั้น
	app.Use("/api", fiberAuth.New(fiberAuth.Config{
		Users: map[string]string{
			"gofiber": "21022566",
		},
		Unauthorized: func(ctx *fiber.Ctx) error {
			return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Unauthorized",
			})
		},
	}))

	v1 := app.Group("/api/v1")
	v1.Get("/fact/:number", c.Factorial)
	v1.Post("/register", c.Register)

	// Route Group ช่วยรวม Route ที่ใช้ Prefix เดียวกันไว้ด้วยกัน
	v3 := app.Group("/api/v3")
	v3.Get("/mean", c.TestParams)

	log.Fatal(app.Listen(":3000"))
}
