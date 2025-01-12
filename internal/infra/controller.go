package infra

import (
	"context"
	"errors"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/tiagods/auth/internal/adapter/database"
	"github.com/tiagods/auth/internal/adapter/web/handler"
	"github.com/tiagods/auth/internal/domain/service"
	"github.com/tiagods/auth/internal/infra/cache"
	sdkDatabase "github.com/tiagods/auth/internal/infra/database"
	"github.com/tiagods/auth/internal/infra/env"
	"github.com/tiagods/auth/internal/infra/logger"
	localMiddleware "github.com/tiagods/auth/internal/infra/middleware"
	"net/http"
	"os"
	"os/signal"
)

func StartApi() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	log := logger.Init()
	defer log.Sync()

	env.GetEnvAsString(env.MYSQL_USER, env.DEFAULT_MYSQL_USER)
	env.GetEnvAsString(env.MYSQL_PASS, env.DEFAULT_MYSQL_PASS)
	env.GetEnvAsString(env.MYSQL_HOST, env.DEFAULT_MYSQL_HOST)
	env.GetEnvAsString(env.MYSQL_DATABASE, env.DEFAULT_MYSQL_DATABASE)

	db := sdkDatabase.NewDB(
		env.GetEnvAsString(env.MYSQL_USER, env.DEFAULT_MYSQL_USER),
		env.GetEnvAsString(env.MYSQL_PASS, env.DEFAULT_MYSQL_PASS),
		env.GetEnvAsString(env.MYSQL_HOST, env.DEFAULT_MYSQL_HOST),
		env.GetEnvAsString(env.MYSQL_DATABASE, env.DEFAULT_MYSQL_DATABASE),
	)
	defer db.Close()

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	//e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
	//	Skipper:      middleware.DefaultSkipper,
	//	ErrorMessage: "Timeout request.",
	//	OnTimeoutRouteErrorHandler: func(err error, c echo.Context) {
	//		c.Logger().Error(c.Path())
	//	},
	//	Timeout: 10 * time.Second,
	//}))

	e.Use(localMiddleware.RequestContext)
	e.HTTPErrorHandler = localMiddleware.HTTPErrorHandler
	e.Validator = localMiddleware.NewValidator()

	repo := database.NewRepository(db, db)
	cacheRepo := cache.NewCache()
	healthService := service.NewHealthService(repo, cacheRepo)

	healthHandler := handler.NewHealthHandler(healthService)
	tokenService := service.NewTokenService(repo, cacheRepo)
	tokenHandler := handler.NewTokenHandler(tokenService)

	e.GET("/health", healthHandler.Health)
	e.POST("/login", tokenHandler.Login)
	e.POST("/register", tokenHandler.Register)
	e.GET("/private", localMiddleware.Private, localMiddleware.IsLoggedIn)
	e.GET("/admin", localMiddleware.Private, localMiddleware.IsLoggedIn, localMiddleware.IsAdmin)
	e.POST("/refresh-token", tokenHandler.RefreshToken)

	go func() {
		if err := e.Start(":8080"); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal(ctx, err, "shutting down the server")
		}
	}()

	<-ctx.Done()
	if err := e.Shutdown(ctx); err != nil {
		logger.Fatal(ctx, err, "shutting down the server")
	}
}
