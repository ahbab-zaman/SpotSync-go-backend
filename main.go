package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/yourusername/spotsync/middleware"
	"github.com/yourusername/spotsync/models"
)

func main() {
	godotenv.Load()

	db := connectDB()

	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	if err := db.AutoMigrate(
		&models.User{},
		&models.ParkingZone{},
		&models.Reservation{},
	); err != nil {
		log.Fatal("Failed to auto-migrate:", err)
	}
	log.Println("Database migration completed")

	e := echo.New()
	e.Validator = &customValidator{}

	e.Use(echomw.Logger())
	e.Use(echomw.Recover())
	e.Use(echomw.CORSWithConfig(echomw.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.DELETE},
		AllowHeaders: []string{echo.HeaderContentType, echo.HeaderAuthorization},
	}))

	api := e.Group("/api/v1")

	api.GET("/protected-test", func(c echo.Context) error {
		userID := c.Get("userID").(uint)
		role := c.Get("role").(string)
		return c.JSON(http.StatusOK, map[string]interface{}{
			"success": true,
			"message": "Authenticated",
			"data": map[string]interface{}{
				"user_id": userID,
				"role":    role,
			},
		})
	}, middleware.JWTMiddleware())

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	e.Logger.Fatal(e.Start(":" + port))
}

func connectDB() *gorm.DB {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	return db
}

type customValidator struct{}

func (cv *customValidator) Validate(i interface{}) error {
	return nil
}
