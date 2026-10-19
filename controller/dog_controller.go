package controller

import (
	"strconv"
	"strings"

	"gofiber-workshop/database"
	"gofiber-workshop/models"

	"github.com/gofiber/fiber/v2"
)

func GetDogs(ctx *fiber.Ctx) error {
	dogs := make([]models.Dogs, 0)
	if err := database.DBConn.Order("id").Find(&dogs).Error; err != nil {
		return databaseError(ctx, err)
	}
	return ctx.JSON(dogs)
}

func GetDog(ctx *fiber.Ctx) error {
	search, err := strconv.Atoi(strings.TrimSpace(ctx.Query("search")))
	if err != nil || search < 0 {
		return apiError(ctx, fiber.StatusBadRequest, "search must be a non-negative integer dog_id")
	}
	dogs := make([]models.Dogs, 0)
	result := database.DBConn.Where("dog_id = ?", search).Order("id").Find(&dogs)
	if result.Error != nil {
		return databaseError(ctx, result.Error)
	}
	if result.RowsAffected == 0 {
		return apiError(ctx, fiber.StatusNotFound, "Dog not found")
	}
	return ctx.JSON(dogs)
}

func AddDog(ctx *fiber.Ctx) error {
	var request models.DogRequest
	if err := ctx.BodyParser(&request); err != nil {
		return apiError(ctx, fiber.StatusBadRequest, "Invalid JSON body")
	}
	request.Name = strings.TrimSpace(request.Name)
	if err := validate.Struct(request); err != nil {
		return apiError(ctx, fiber.StatusBadRequest, "name is required (max 100 characters); dog_id is required and must be a non-negative integer")
	}
	dog := models.Dogs{Name: request.Name, DogID: *request.DogID}
	if err := database.DBConn.Create(&dog).Error; err != nil {
		return databaseError(ctx, err)
	}
	return ctx.Status(fiber.StatusCreated).JSON(dog)
}

func UpdateDog(ctx *fiber.Ctx) error {
	id, err := parseID(ctx.Params("id"))
	if err != nil {
		return apiError(ctx, fiber.StatusBadRequest, err.Error())
	}
	var request models.DogRequest
	if err := ctx.BodyParser(&request); err != nil {
		return apiError(ctx, fiber.StatusBadRequest, "Invalid JSON body")
	}
	request.Name = strings.TrimSpace(request.Name)
	if err := validate.Struct(request); err != nil {
		return apiError(ctx, fiber.StatusBadRequest, "name is required (max 100 characters); dog_id is required and must be a non-negative integer")
	}
	var dog models.Dogs
	if err := database.DBConn.First(&dog, id).Error; err != nil {
		return recordError(ctx, err, "Dog")
	}
	// Map อัปเดต dog_id เป็น 0 ได้ และรับเฉพาะ Field ที่อนุญาต
	result := database.DBConn.Model(&dog).Updates(map[string]interface{}{
		"name": request.Name, "dog_id": *request.DogID,
	})
	if result.Error != nil {
		return databaseError(ctx, result.Error)
	}
	if result.RowsAffected == 0 {
		// ตรวจซ้ำเพื่อแยกกรณีค่าไม่เปลี่ยนกับรายการถูกลบระหว่างแก้ไข
		if err := database.DBConn.First(&dog, id).Error; err != nil {
			return recordError(ctx, err, "Dog")
		}
	}
	return ctx.JSON(dog)
}

func RemoveDog(ctx *fiber.Ctx) error {
	id, err := parseID(ctx.Params("id"))
	if err != nil {
		return apiError(ctx, fiber.StatusBadRequest, err.Error())
	}
	// Delete ร่วมกับ gorm.Model จะตั้ง deleted_at โดยไม่ลบแถวจริง
	result := database.DBConn.Delete(&models.Dogs{}, id)
	if result.Error != nil {
		return databaseError(ctx, result.Error)
	}
	if result.RowsAffected == 0 {
		return apiError(ctx, fiber.StatusNotFound, "Dog not found")
	}
	return ctx.JSON(fiber.Map{"success": true, "message": "Dog deleted"})
}

func GetDeletedDogs(ctx *fiber.Ctx) error {
	dogs := make([]models.Dogs, 0)
	if err := database.DBConn.Unscoped().Where("deleted_at IS NOT NULL").Order("id").Find(&dogs).Error; err != nil {
		return databaseError(ctx, err)
	}
	return ctx.JSON(fiber.Map{"count": len(dogs), "data": dogs})
}

func GetDogsByRange(ctx *fiber.Ctx) error {
	dogs := make([]models.Dogs, 0)
	// ใช้ dog_id ของโจทย์ ไม่ใช่ Primary Key id และไม่รวม Soft Delete
	if err := database.DBConn.Where("dog_id > ? AND dog_id < ?", 50, 100).Order("id").Find(&dogs).Error; err != nil {
		return databaseError(ctx, err)
	}
	return ctx.JSON(fiber.Map{"count": len(dogs), "data": dogs})
}

func GetDogsJson(ctx *fiber.Ctx) error {
	dogs := make([]models.Dogs, 0)
	if err := database.DBConn.Order("id").Find(&dogs).Error; err != nil {
		return databaseError(ctx, err)
	}
	return ctx.JSON(buildDogSummary(dogs))
}

func dogType(dogID int) string {
	switch {
	case dogID >= 10 && dogID <= 50:
		return "red"
	case dogID >= 100 && dogID <= 150:
		return "green"
	case dogID >= 200 && dogID <= 250:
		return "pink"
	default:
		return "no color"
	}
}

// แยกการคำนวณออกจาก Query เพื่อทดสอบได้โดยไม่ต้องใช้ MySQL
func buildDogSummary(dogs []models.Dogs) models.ResultData {
	data := make([]models.DogsRes, 0, len(dogs))
	sumRed, sumGreen, sumPink, sumNoColor := 0, 0, 0, 0
	for _, dog := range dogs {
		typeStr := dogType(dog.DogID)
		switch typeStr {
		case "red":
			sumRed++
		case "green":
			sumGreen++
		case "pink":
			sumPink++
		default:
			sumNoColor++
		}
		data = append(data, models.DogsRes{Name: dog.Name, DogID: dog.DogID, Type: typeStr})
	}
	return models.ResultData{
		Data: data, Name: "golang-test", Count: len(data),
		SumRed: sumRed, SumGreen: sumGreen, SumPink: sumPink, SumNoColor: sumNoColor,
	}
}
