package infra

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/tiagods/auth/internal/adapter/database"
	"github.com/tiagods/auth/internal/adapter/web/handler"
	"github.com/tiagods/auth/internal/domain"
	service "github.com/tiagods/auth/internal/domain/services"
	"github.com/tiagods/auth/internal/infra/cache"
	"github.com/tiagods/auth/internal/infra/database/mysql"
	tokenMiddleware "github.com/tiagods/auth/internal/infra/middleware"
	"time"
)

func StartApi() {
	db := mysql.NewMysqlDB()
	defer db.Close()

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Skipper:      middleware.DefaultSkipper,
		ErrorMessage: "Timeout request.",
		OnTimeoutRouteErrorHandler: func(err error, c echo.Context) {
			c.Logger().Error(c.Path())
		},
		Timeout: 30 * time.Second,
	}))
	e.Validator = tokenMiddleware.NewValidator()

	repo := database.NewDatabaseRepository(db)
	memory := cache.NewMemoryCache()
	tokenService := service.NewTokenService(repo, memory)

	m := domain.NewManager(tokenService)
	e.GET("/health", handler.Health)

	e.POST("/login", handler.Login)
	e.POST("/register", handler.Register)
	e.GET("/private", tokenMiddleware.Private, tokenMiddleware.IsLoggedIn)
	e.GET("/admin", tokenMiddleware.Private, tokenMiddleware.IsLoggedIn, tokenMiddleware.IsAdmin)
	e.POST("/refresh-token", handler.RefreshToken)

	e.Logger.Fatal(e.Start(":8080"))
}
