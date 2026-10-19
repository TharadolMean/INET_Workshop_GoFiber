package controller

import (
	"strings"

	"gofiber-workshop/database"
	"gofiber-workshop/models"

	"github.com/gofiber/fiber/v2"
)

func GetCompanies(ctx *fiber.Ctx) error {
	companies := make([]models.Company, 0)
	if err := database.DBConn.Order("id").Find(&companies).Error; err != nil {
		return databaseError(ctx, err)
	}
	return ctx.JSON(fiber.Map{"count": len(companies), "data": companies})
}

func GetCompany(ctx *fiber.Ctx) error {
	id, err := parseID(ctx.Params("id"))
	if err != nil {
		return apiError(ctx, fiber.StatusBadRequest, err.Error())
	}
	var company models.Company
	if err := database.DBConn.First(&company, id).Error; err != nil {
		return recordError(ctx, err, "Company")
	}
	return ctx.JSON(company)
}

func AddCompany(ctx *fiber.Ctx) error {
	var request models.CompanyRequest
	if err := ctx.BodyParser(&request); err != nil {
		return apiError(ctx, fiber.StatusBadRequest, "Invalid JSON body")
	}
	trimCompany(&request)
	if err := validate.Struct(request); err != nil {
		return apiError(ctx, fiber.StatusBadRequest, "name is required (max 100); address max 255; phone max 30; email must be valid (max 254)")
	}
	company := models.Company{Name: request.Name, Address: request.Address, Phone: request.Phone, Email: request.Email}
	if err := database.DBConn.Create(&company).Error; err != nil {
		return databaseError(ctx, err)
	}
	return ctx.Status(fiber.StatusCreated).JSON(company)
}

func UpdateCompany(ctx *fiber.Ctx) error {
	id, err := parseID(ctx.Params("id"))
	if err != nil {
		return apiError(ctx, fiber.StatusBadRequest, err.Error())
	}
	var request models.CompanyRequest
	if err := ctx.BodyParser(&request); err != nil {
		return apiError(ctx, fiber.StatusBadRequest, "Invalid JSON body")
	}
	trimCompany(&request)
	if err := validate.Struct(request); err != nil {
		return apiError(ctx, fiber.StatusBadRequest, "name is required (max 100); address max 255; phone max 30; email must be valid (max 254)")
	}
	var company models.Company
	if err := database.DBConn.First(&company, id).Error; err != nil {
		return recordError(ctx, err, "Company")
	}
	// PUT รับข้อมูลชุดใหม่ทั้งหมด; Map ทำให้ล้าง Optional Field เป็นค่าว่างได้
	result := database.DBConn.Model(&company).Updates(map[string]interface{}{
		"name": request.Name, "address": request.Address, "phone": request.Phone, "email": request.Email,
	})
	if result.Error != nil {
		return databaseError(ctx, result.Error)
	}
	if result.RowsAffected == 0 {
		if err := database.DBConn.First(&company, id).Error; err != nil {
			return recordError(ctx, err, "Company")
		}
	}
	return ctx.JSON(company)
}

func RemoveCompany(ctx *fiber.Ctx) error {
	id, err := parseID(ctx.Params("id"))
	if err != nil {
		return apiError(ctx, fiber.StatusBadRequest, err.Error())
	}
	result := database.DBConn.Delete(&models.Company{}, id)
	if result.Error != nil {
		return databaseError(ctx, result.Error)
	}
	if result.RowsAffected == 0 {
		return apiError(ctx, fiber.StatusNotFound, "Company not found")
	}
	return ctx.JSON(fiber.Map{"success": true, "message": "Company deleted"})
}

func trimCompany(request *models.CompanyRequest) {
	request.Name = strings.TrimSpace(request.Name)
	request.Address = strings.TrimSpace(request.Address)
	request.Phone = strings.TrimSpace(request.Phone)
	request.Email = strings.TrimSpace(request.Email)
}
