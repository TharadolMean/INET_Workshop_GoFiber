package main

import (
	"log"

	// Alias c ทำให้เรียก Handler ใน Package controller ด้วยชื่อสั้น เช่น c.TestParams
	c "gofiber-workshop/controller"
	"gofiber-workshop/database"

	"github.com/gofiber/fiber/v2"
	fiberAuth "github.com/gofiber/fiber/v2/middleware/basicauth"
)

func main() {
	if err := database.Connect(); err != nil {
		log.Printf("Database unavailable: %v. Exercise 5–6 remain available; restart after configuring MySQL.", err)
	} else {
		log.Print("Database connected and migrated")
		sqlDB, err := database.DBConn.DB()
		if err != nil {
			log.Fatal(err)
		}
		defer sqlDB.Close()
	}
	app := newApp()
	if err := app.Listen(":3000"); err != nil {
		log.Fatal(err)
	}
}

// แยกการกำหนด Routes เพื่อทดสอบ HTTP ได้โดยไม่ต้องเปิด Port จริง
func newApp() *fiber.App {
	app := fiber.New()

	// ข้อ 5.0: กำหนด Basic Auth ให้ทุก Route ที่อยู่ใต้ /api
	// Username: gofiber, Password: 21022566
	// Basic Auth ทำงานก่อนเข้า Handler ทุก Route ที่อยู่ใต้ /api
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

	// ข้อ 5.1 และข้อ 6: สร้าง Route Group /api/v1
	v1 := app.Group("/api/v1")
	// ข้อ 5.1: รับค่า number ผ่าน Path Parameter แล้วคำนวณ Factorial
	v1.Get("/fact/:number", c.Factorial)
	// ข้อ 6: POST สมัครสมาชิกและตรวจสอบความถูกต้องของข้อมูล
	v1.Post("/register", c.Register)

	dog := v1.Group("/dog", c.RequireDatabase)
	dog.Get("", c.GetDogs)
	dog.Get("/filter", c.GetDog)
	dog.Get("/json", c.GetDogsJson)
	dog.Get("/deleted", c.GetDeletedDogs)
	dog.Get("/range", c.GetDogsByRange)
	dog.Post("/", c.AddDog)
	dog.Put("/:id", c.UpdateDog)
	dog.Delete("/:id", c.RemoveDog)

	company := v1.Group("/company", c.RequireDatabase)
	company.Post("", c.AddCompany)
	company.Get("", c.GetCompanies)
	company.Get("/:id", c.GetCompany)
	company.Put("/:id", c.UpdateCompany)
	company.Delete("/:id", c.RemoveCompany)

	// ข้อ 5.2: สร้าง Route Group /api/v3 และรับค่า tax_id ผ่าน Query Parameter
	// Route Group ช่วยรวม Route ที่ใช้ Prefix เดียวกันไว้ด้วยกัน
	v3 := app.Group("/api/v3")
	v3.Get("/mean", c.TestParams)

	return app
}
