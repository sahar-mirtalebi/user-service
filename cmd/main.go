package main

import (
	"log"
	"user-service/auth"
	"user-service/user"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewLogger() (*zap.Logger, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}
	return logger, nil
}

func NewDB() (*gorm.DB, error) {
	dsn := "host=localhost user=admin password=sahar223010 dbname=rental_service_db port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

func NewValidator() *validator.Validate {
	return validator.New()
}

func RegisterRoutes(e *echo.Echo, handler *user.UserHandler) {
	e.POST("/register", handler.RegisterUser)
	e.POST("/login", handler.LoginUser)
	e.POST("/forgot-password", handler.FogotPassword)
	e.POST("/reset-password", handler.ResetPassword)

	authGroup := e.Group("/auth")
	authGroup.Use(auth.AuthMiddleware)

	authGroup.DELETE("/users/me", handler.DeleteAccount)
	authGroup.PUT("/users/me", handler.UpdateAccount)
	authGroup.GET("/users/me", handler.RetrieveAccount)

}

func main() {
	e := echo.New()

	app := fx.New(
		fx.Provide(
			NewDB,
			NewLogger,
			NewValidator,
			user.NewRepository,
			user.NewUserService,
			user.NewUserHandler,
			func() *echo.Echo { return e },
		),
		fx.Invoke(
			func(e *echo.Echo, handler *user.UserHandler) {
				RegisterRoutes(e, handler)
			},
			func() {
				if err := e.Start(":8080"); err != nil {
					log.Fatal("Echo server failed to start", zap.Error(err))
				}
			},
		),
	)
	app.Run()
}
