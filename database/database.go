package database

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"gofiber-workshop/models"

	sqlmysql "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DBConn *gorm.DB

func Connect() error {
	// Environment Variables ที่ตั้งไว้มีลำดับเหนือค่าใน .env
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("cannot load .env; check file syntax and permissions")
	}
	user := os.Getenv("DB_USER")
	if user == "" {
		return fmt.Errorf("DB_USER is required; see .env.example and README.md")
	}
	port := envOrDefault("DB_PORT", "3306")
	portNumber, err := strconv.Atoi(port)
	if err != nil || portNumber < 1 || portNumber > 65535 {
		return fmt.Errorf("DB_PORT must be between 1 and 65535")
	}

	// FormatDSN จัดรูปแบบค่าต่าง ๆ โดยไม่พิมพ์รหัสผ่านลง Log
	config := sqlmysql.NewConfig()
	config.User = user
	config.Passwd = os.Getenv("DB_PASSWORD")
	config.Net = "tcp"
	config.Addr = net.JoinHostPort(envOrDefault("DB_HOST", "127.0.0.1"), port)
	config.DBName = envOrDefault("DB_NAME", "golang_test")
	config.ParseTime = true
	config.Loc = time.Local
	config.Params = map[string]string{"charset": "utf8mb4"}
	config.Timeout = 5 * time.Second
	config.ReadTimeout = 5 * time.Second
	config.WriteTimeout = 5 * time.Second

	db, err := gorm.Open(mysql.Open(config.FormatDSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return fmt.Errorf("connect MySQL: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get database connection: %w", err)
	}
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)
	if err := db.AutoMigrate(&models.Dogs{}, &models.Company{}); err != nil {
		_ = sqlDB.Close()
		return fmt.Errorf("AutoMigrate: %w", err)
	}
	// เปิดให้ Handler ใช้งานหลังเชื่อมต่อและ Migrate สำเร็จเท่านั้น
	DBConn = db
	return nil
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
