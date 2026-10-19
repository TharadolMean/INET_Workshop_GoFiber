package controller

import (
	"regexp"
	"strconv"
	"unicode/utf8"

	"gofiber-workshop/models"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

var validate = newValidator()

func newValidator() *validator.Validate {
	v := validator.New()

	_ = v.RegisterValidation("username", func(fl validator.FieldLevel) bool {
		matched, _ := regexp.MatchString(`^[A-Za-z0-9_-]+$`, fl.Field().String())
		return matched
	})
	_ = v.RegisterValidation("phone", func(fl validator.FieldLevel) bool {
		matched, _ := regexp.MatchString(`^0[0-9]{9}$`, fl.Field().String())
		return matched
	})
	_ = v.RegisterValidation("website", func(fl validator.FieldLevel) bool {
		matched, _ := regexp.MatchString(`^[A-Za-z0-9-]+$`, fl.Field().String())
		return matched
	})

	return v
}

// ข้อ 5.1: รับ Path Parameter จาก /api/v1/fact/:number
// แล้วคำนวณและคืนค่า Factorial ของตัวเลขนั้น
// Factorial รับ Path Parameter ซึ่งเป็นค่าที่อยู่ในส่วนเส้นทางของ URL
func Factorial(ctx *fiber.Ctx) error {
	// Params อ่านค่าที่แทน :number จาก Route /fact/:number
	numberText := ctx.Params("number", "5")

	// URL เป็นข้อความ จึงแปลงเป็น int ก่อนนำไปคำนวณ
	number, err := strconv.Atoi(numberText)
	if err != nil || number < 0 || number > 20 {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "number must be an integer from 0 to 20",
		})
	}

	// เริ่มผลคูณที่ 1 แล้วคูณเลขตั้งแต่ 1 ถึง number ด้วย for Loop
	factorial := 1
	for i := 1; i <= number; i++ {
		factorial *= i
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"number":    number,
		"factorial": factorial,
	})
}

// ข้อ 5.2: รับ Query Parameter ชื่อ tax_id จาก /api/v3/mean
// แล้วแปลงตัวอักษรแต่ละตัวเป็นรหัส ASCII
// TestParams แปลง Query Parameter tax_id เป็นรหัส ASCII ทีละตัว
func TestParams(ctx *fiber.Ctx) error {
	// Query Parameter อยู่หลังเครื่องหมาย ? ส่วน Params อยู่ใน Path ของ URL
	taxID := ctx.Query("tax_id")
	if taxID == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "tax_id is required",
		})
	}

	// วนตาม Unicode rune เพื่อปฏิเสธอักขระที่ไม่ใช่ ASCII อย่างชัดเจน
	// ตรวจความถูกต้องของ UTF-8 ก่อน เพื่อไม่ให้ byte ที่ผิดรูปถูกแปลงผิดความหมาย
	if !utf8.ValidString(taxID) {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "tax_id must contain ASCII characters only",
		})
	}

	asciiCodes := make([]int, 0, len(taxID))
	for _, character := range taxID {
		if character > 127 {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "tax_id must contain ASCII characters only",
			})
		}
		asciiCodes = append(asciiCodes, int(character))
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"tax_id": taxID,
		"ascii":  asciiCodes,
	})
}

// ข้อ 6: รับข้อมูลสมัครสมาชิกด้วย POST /api/v1/register
// ตรวจสอบ JSON และข้อมูลด้วย validator/v10 ก่อนส่งผลลัพธ์กลับ
// Register อ่าน JSON, ตรวจ Struct ด้วย validator/v10 แล้วส่งกลับเฉพาะข้อมูลที่เปิดเผยได้
func Register(ctx *fiber.Ctx) error {
	var request models.RegisterRequest

	// BodyParser แปลง JSON Request Body ให้เป็นค่าใน Struct ตาม JSON Tags
	if err := ctx.BodyParser(&request); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid JSON body",
		})
	}

	// Validator อ่านกฎจาก validate Tags และคืนข้อผิดพลาดของหลาย Field ได้ในครั้งเดียว
	if err := validate.Struct(request); err != nil {
		validationErrors := make(map[string]string)
		if errors, ok := err.(validator.ValidationErrors); ok {
			for _, fieldError := range errors {
				fieldName := jsonFieldName(fieldError.Field())
				// เก็บข้อผิดพลาดแรกของแต่ละ Field เพื่อให้ required ชัดเจนกว่ากฎถัดไป
				if _, alreadyAdded := validationErrors[fieldName]; !alreadyAdded {
					validationErrors[fieldName] = validationMessage(fieldName, fieldError.Tag())
				}
			}
		}

		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Validation failed",
			"errors":  validationErrors,
		})
	}

	// ไม่ส่ง request.Password กลับไปใน Response
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Validation passed",
		"data": fiber.Map{
			"email":    request.Email,
			"username": request.Username,
		},
	})
}

func jsonFieldName(field string) string {
	fieldNames := map[string]string{
		"Email":        "email",
		"Username":     "username",
		"Password":     "password",
		"LineID":       "line_id",
		"Phone":        "phone",
		"BusinessType": "business_type",
		"WebsiteName":  "website_name",
	}
	return fieldNames[field]
}

func validationMessage(field string, tag string) string {
	switch field {
	case "email":
		if tag == "required" {
			return "Email is required"
		}
		return "Invalid email format"
	case "username":
		if tag == "required" {
			return "Username is required"
		}
		return "Invalid username format"
	case "password":
		return "Password must be 6-20 characters"
	case "phone":
		if tag == "required" {
			return "Phone is required"
		}
		return "Invalid phone number"
	case "business_type":
		return "Business type is required"
	case "website_name":
		if tag == "required" {
			return "Website name is required"
		}
		return "Invalid website name"
	default:
		return "Invalid value"
	}
}

//เทส
