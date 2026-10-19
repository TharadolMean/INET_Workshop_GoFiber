package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"gofiber-workshop/database"
	"gofiber-workshop/models"
)

// เปิดด้วย MYSQL_INTEGRATION_TEST=1 เท่านั้น; ข้อมูลทดสอบทั้งหมดอยู่ใน Transaction
func TestMySQLIntegration(t *testing.T) {
	if os.Getenv("MYSQL_INTEGRATION_TEST") != "1" {
		t.Skip("set MYSQL_INTEGRATION_TEST=1 and configure MySQL to run database integration tests")
	}
	previous := database.DBConn
	if err := database.Connect(); err != nil {
		t.Fatal(err)
	}
	connection := database.DBConn
	sqlDB, err := connection.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.DBConn = previous; _ = sqlDB.Close() })
	tx := connection.Begin()
	if tx.Error != nil {
		t.Fatal(tx.Error)
	}
	rolledBack := false
	t.Cleanup(func() {
		if !rolledBack {
			if err := tx.Rollback().Error; err != nil {
				t.Errorf("rollback test data: %v", err)
			}
		}
	})
	database.DBConn = tx
	app := newApp()

	call := func(method, path, body string, expected int) []byte {
		t.Helper()
		status, data := performRequest(t, app, method, path, body, workshopAuth)
		if status != expected {
			t.Fatalf("%s %s: got %d, want %d: %s", method, path, status, expected, data)
		}
		return data
	}
	decode := func(body []byte, target interface{}) {
		t.Helper()
		if err := json.Unmarshal(body, target); err != nil {
			t.Fatal(err)
		}
	}
	var baseline models.ResultData
	decode(call("GET", "/api/v1/dog/json", "", 200), &baseline)
	prefix := fmt.Sprintf("IT_%d_", time.Now().UnixNano())
	ids := []int{9, 10, 50, 51, 99, 100, 150, 151, 200, 250, 251, 999}
	created := make([]models.Dogs, 0, len(ids))
	for i, id := range ids {
		// ID และ DeletedAt จาก Client ต้องไม่มีผลต่อ Model ที่สร้าง
		body := fmt.Sprintf(`{"name":%q,"dog_id":%d,"ID":4294967295,"DeletedAt":"2020-01-01T00:00:00Z"}`, fmt.Sprintf("%sDog%02d", prefix, i+1), id)
		var dog models.Dogs
		decode(call("POST", "/api/v1/dog/", body, 201), &dog)
		if dog.ID == 0 || dog.ID == 4294967295 || dog.DeletedAt.Valid {
			t.Fatalf("client modified GORM-managed fields: %+v", dog)
		}
		created = append(created, dog)
	}
	var summary models.ResultData
	decode(call("GET", "/api/v1/dog/json", "", 200), &summary)
	if summary.Count != baseline.Count+12 || len(summary.Data) != summary.Count || summary.Name != "golang-test" ||
		summary.SumRed != baseline.SumRed+2 || summary.SumGreen != baseline.SumGreen+2 ||
		summary.SumPink != baseline.SumPink+2 || summary.SumNoColor != baseline.SumNoColor+6 ||
		summary.SumRed+summary.SumGreen+summary.SumPink+summary.SumNoColor != summary.Count {
		t.Fatalf("incorrect summary delta: before=%+v, after=%+v", baseline, summary)
	}
	var all []models.Dogs
	decode(call("GET", "/api/v1/dog", "", 200), &all)
	found := false
	for _, dog := range all {
		if dog.ID == created[0].ID {
			found = true
		}
	}
	if !found {
		t.Fatal("GET all missing created dog")
	}
	var filtered []models.Dogs
	decode(call("GET", "/api/v1/dog/filter?search=10", "", 200), &filtered)
	found = false
	for _, dog := range filtered {
		if dog.DogID != 10 {
			t.Fatalf("filter used wrong field: %+v", dog)
		}
		if dog.ID == created[1].ID {
			found = true
		}
	}
	if !found {
		t.Fatal("filter missing created DogID=10")
	}
	var ranged struct {
		Count int           `json:"count"`
		Data  []models.Dogs `json:"data"`
	}
	decode(call("GET", "/api/v1/dog/range", "", 200), &ranged)
	if ranged.Count != len(ranged.Data) {
		t.Fatal("range count mismatch")
	}
	found51, found99 := false, false
	for _, dog := range ranged.Data {
		if dog.DogID <= 50 || dog.DogID >= 100 || dog.DeletedAt.Valid {
			t.Fatalf("range includes excluded dog: %+v", dog)
		}
		found51 = found51 || dog.ID == created[3].ID
		found99 = found99 || dog.ID == created[4].ID
	}
	if !found51 || !found99 {
		t.Fatal("range missing boundary fixture 51 or 99")
	}

	dogPath := fmt.Sprintf("/api/v1/dog/%d", created[3].ID)
	var updated models.Dogs
	decode(call("PUT", dogPath, `{"name":"Updated Dog","dog_id":0,"ID":4294967295,"DeletedAt":"2020-01-01T00:00:00Z"}`, 200), &updated)
	if updated.ID != created[3].ID || updated.DogID != 0 || updated.Name != "Updated Dog" || updated.DeletedAt.Valid {
		t.Fatalf("incorrect update or unsafe fields: %+v", updated)
	}
	call("DELETE", dogPath, "", 200)
	call("DELETE", dogPath, "", 404)
	call("PUT", dogPath, `{"name":"Missing","dog_id":1}`, 404)
	var deleted struct {
		Count int           `json:"count"`
		Data  []models.Dogs `json:"data"`
	}
	decode(call("GET", "/api/v1/dog/deleted", "", 200), &deleted)
	if deleted.Count != len(deleted.Data) {
		t.Fatal("deleted count mismatch")
	}
	found = false
	for _, dog := range deleted.Data {
		if !dog.DeletedAt.Valid {
			t.Fatalf("deleted endpoint returned active dog: %+v", dog)
		}
		found = found || dog.ID == created[3].ID
	}
	if !found {
		t.Fatal("soft-deleted fixture missing")
	}
	var stored models.Dogs
	if err := tx.Unscoped().First(&stored, created[3].ID).Error; err != nil || !stored.DeletedAt.Valid {
		t.Fatalf("delete must retain row with deleted_at: %+v %v", stored, err)
	}
	decode(call("GET", "/api/v1/dog", "", 200), &all)
	for _, dog := range all {
		if dog.ID == created[3].ID || dog.DeletedAt.Valid {
			t.Fatal("regular GET includes soft-deleted data")
		}
	}

	var company models.Company
	decode(call("POST", "/api/v1/company", `{"name":"Integration Company","address":"Bangkok","phone":"021234567","email":"test@example.com","ID":4294967295}`, 201), &company)
	if company.ID == 0 || company.ID == 4294967295 {
		t.Fatalf("invalid server-generated company ID: %d", company.ID)
	}
	companyPath := fmt.Sprintf("/api/v1/company/%d", company.ID)
	call("GET", companyPath, "", 200)
	var companies struct {
		Count int              `json:"count"`
		Data  []models.Company `json:"data"`
	}
	decode(call("GET", "/api/v1/company", "", 200), &companies)
	found = false
	for _, item := range companies.Data {
		found = found || item.ID == company.ID
	}
	if companies.Count != len(companies.Data) || !found {
		t.Fatal("company list missing created company or incorrect count")
	}
	decode(call("PUT", companyPath, `{"name":"Updated Company","address":"","phone":"","email":"","DeletedAt":"2020-01-01T00:00:00Z"}`, 200), &company)
	if company.Name != "Updated Company" || company.Email != "" || company.Phone != "" || company.Address != "" || company.DeletedAt.Valid {
		t.Fatalf("PUT did not clear optional fields safely: %+v", company)
	}
	call("DELETE", companyPath, "", 200)
	call("GET", companyPath, "", 404)
	call("PUT", companyPath, `{"name":"Missing"}`, 404)
	call("DELETE", companyPath, "", 404)
	if err := tx.Unscoped().First(&company, company.ID).Error; err != nil || !company.DeletedAt.Valid {
		t.Fatalf("company delete was not soft delete: %v", err)
	}

	for _, tc := range []struct{ method, path, body string }{
		{"POST", "/api/v1/dog/", `{}`},
		{"POST", "/api/v1/dog/", `{"name":" ","dog_id":1}`},
		{"POST", "/api/v1/dog/", `{"name":"Dog","dog_id":-1}`},
		{"POST", "/api/v1/dog/", `{"name":`},
		{"PUT", dogPath, `{"name":"Dog"}`},
		{"DELETE", "/api/v1/dog/invalid", ""},
		{"PUT", "/api/v1/dog/0", `{"name":"Dog","dog_id":1}`},
		{"GET", "/api/v1/dog/filter?search=abc", ""},
		{"GET", "/api/v1/dog/filter", ""},
		{"GET", "/api/v1/company/0", ""},
		{"DELETE", "/api/v1/company/-1", ""},
		{"PUT", "/api/v1/company/abc", `{"name":"Company"}`},
		{"POST", "/api/v1/company", `{}`},
		{"POST", "/api/v1/company", `{"name":"Company","email":"invalid"}`},
	} {
		call(tc.method, tc.path, tc.body, 400)
	}
	// ตรวจ Query error จริงโดยใช้ Connection ที่ปิดแล้ว หลัง Rollback ข้อมูลทดสอบ
	if err := tx.Rollback().Error; err != nil {
		t.Fatal(err)
	}
	rolledBack = true
	database.DBConn = connection
	if err := sqlDB.Close(); err != nil {
		t.Fatal(err)
	}
	status, body := performRequest(t, app, "GET", "/api/v1/dog/json", "", workshopAuth)
	if status != 500 || !strings.Contains(string(body), "Database operation failed") {
		t.Fatalf("closed DB must return JSON 500: %d %s", status, body)
	}
}
