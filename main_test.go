package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"gofiber-workshop/database"

	"github.com/gofiber/fiber/v2"
)

const workshopAuth = "gofiber:21022566"

func performRequest(t *testing.T, app *fiber.App, method, path, body, auth string) (int, []byte) {
	t.Helper()
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	if auth != "" {
		request.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(auth)))
	}
	response, err := app.Test(request, 10000)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	content, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(response.Header.Get("Content-Type"), "application/json") || !json.Valid(content) {
		t.Fatalf("expected JSON, got %s: %s", response.Header.Get("Content-Type"), content)
	}
	return response.StatusCode, content
}

func TestExistingExercises(t *testing.T) {
	app := newApp()
	validRegistration := `{"email":"test@example.com","username":"mean123","password":"Test123456","phone":"0812345678","business_type":"Technology","website_name":"myshop"}`
	cases := []struct {
		method, path, body string
		status             int
	}{
		{"GET", "/api/v1/fact/5", "", 200},
		{"GET", "/api/v1/fact/0", "", 200},
		{"GET", "/api/v1/fact/20", "", 200},
		{"GET", "/api/v1/fact/21", "", 400},
		{"GET", "/api/v1/fact/-1", "", 400},
		{"GET", "/api/v1/fact/abc", "", 400},
		{"GET", "/api/v3/mean?tax_id=Hello", "", 200},
		{"GET", "/api/v3/mean?tax_id=123", "", 200},
		{"GET", "/api/v3/mean", "", 400},
		{"GET", "/api/v3/mean?tax_id=%E0%B9%84%E0%B8%97%E0%B8%A2", "", 400},
		{"POST", "/api/v1/register", validRegistration, 200},
		{"POST", "/api/v1/register", `{}`, 400},
		{"POST", "/api/v1/register", `{"email":`, 400},
	}
	for i, tc := range cases {
		t.Run(fmt.Sprintf("%02d %s %s", i, tc.method, tc.path), func(t *testing.T) {
			status, content := performRequest(t, app, tc.method, tc.path, tc.body, workshopAuth)
			if status != tc.status {
				t.Fatalf("status %d, want %d: %s", status, tc.status, content)
			}
			var result map[string]interface{}
			if err := json.Unmarshal(content, &result); err != nil {
				t.Fatal(err)
			}
			if tc.path == "/api/v1/fact/5" && result["factorial"] != float64(120) {
				t.Fatalf("incorrect factorial: %s", content)
			}
			if tc.path == "/api/v3/mean?tax_id=Hello" {
				want := []int{72, 101, 108, 108, 111}
				got := result["ascii"].([]interface{})
				if len(got) != len(want) {
					t.Fatalf("incorrect ASCII length: %s", content)
				}
				for i, code := range want {
					if got[i] != float64(code) {
						t.Fatalf("incorrect ASCII: %s", content)
					}
				}
			}
			if tc.method == "POST" && status == 200 && (strings.Contains(string(content), "Test123456") || strings.Contains(string(content), `"password"`)) {
				t.Fatalf("password exposed: %s", content)
			}
		})
	}
}

func TestBasicAuthCoversOldAndNewRoutes(t *testing.T) {
	app := newApp()
	for _, path := range []string{"/api/v1/fact/5", "/api/v3/mean?tax_id=Hello", "/api/v1/register", "/api/v1/dog", "/api/v1/dog/json", "/api/v1/company"} {
		for _, auth := range []string{"", "gofiber:wrong"} {
			status, body := performRequest(t, app, "GET", path, "", auth)
			if status != 401 || !strings.Contains(string(body), "Unauthorized") {
				t.Fatalf("%s: expected JSON 401, got %d %s", path, status, body)
			}
		}
	}
}

func TestDatabaseUnavailableRoutes(t *testing.T) {
	previous := database.DBConn
	database.DBConn = nil
	t.Cleanup(func() { database.DBConn = previous })
	app := newApp()
	for _, tc := range []struct{ method, path string }{
		{"GET", "/api/v1/dog"}, {"GET", "/api/v1/dog/filter?search=10"},
		{"GET", "/api/v1/dog/json"}, {"GET", "/api/v1/dog/deleted"}, {"GET", "/api/v1/dog/range"},
		{"POST", "/api/v1/dog/"}, {"PUT", "/api/v1/dog/1"}, {"DELETE", "/api/v1/dog/1"},
		{"GET", "/api/v1/company"}, {"GET", "/api/v1/company/1"},
		{"POST", "/api/v1/company"}, {"PUT", "/api/v1/company/1"}, {"DELETE", "/api/v1/company/1"},
	} {
		status, body := performRequest(t, app, tc.method, tc.path, "", workshopAuth)
		if status != 500 || !strings.Contains(string(body), "Database unavailable") {
			t.Fatalf("%s %s: expected controlled JSON 500, got %d %s", tc.method, tc.path, status, body)
		}
	}
}
