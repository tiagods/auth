package infra

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

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
	"github.com/tiagods/auth/internal/infra/tracer"
)

func StartApi() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	log := logger.Init()
	defer log.Sync()

	otelShutdown, err := tracer.SetupOTelSDK(ctx)
	if err != nil {
		logger.Fatal(ctx, err, "failed to initialize opentelemetry")
	}
	defer otelShutdown(ctx)

	env.GetEnvAsString(env.MYSQL_USER, env.DEFAULT_MYSQL_USER)
	env.GetEnvAsString(env.MYSQL_PASS, env.DEFAULT_MYSQL_PASS)
	env.GetEnvAsString(env.MYSQL_HOST, env.DEFAULT_MYSQL_HOST)
	env.GetEnvAsString(env.MYSQL_DATABASE, env.DEFAULT_MYSQL_DATABASE)

	db, err := sdkDatabase.NewDB(
		env.GetEnvAsString(env.MYSQL_USER, env.DEFAULT_MYSQL_USER),
		env.GetEnvAsString(env.MYSQL_PASS, env.DEFAULT_MYSQL_PASS),
		env.GetEnvAsString(env.MYSQL_HOST, env.DEFAULT_MYSQL_HOST),
		env.GetEnvAsString(env.MYSQL_DATABASE, env.DEFAULT_MYSQL_DATABASE),
	)
	if err != nil {
		logger.Fatal(ctx, err, "failed to initialize database")
	}

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
	//	Timeout: 60 * time.Second,
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

	handlerTest := func(c echo.Context) error {
		ctx := c.Request().Context()
		logger.Info(ctx, "Test handler called")
		return c.String(http.StatusOK, "Test successful")
	}

	e.GET("/health", healthHandler.Health)
	e.POST("/login", tokenHandler.Login)
	e.POST("/register", tokenHandler.Register)
	e.GET("/private", localMiddleware.Private, localMiddleware.IsLoggedIn)
	e.GET("/admin", localMiddleware.Private, localMiddleware.IsLoggedIn, localMiddleware.IsAdmin)
	e.POST("/token", tokenHandler.RecreateToken)
	e.DELETE("/revoke-token", tokenHandler.RevokeToken)
	e.GET("/validate-token", tokenHandler.ValidateToken) //localMiddleware.IsLoggedIn)
	e.GET("/test", handlerTest, localMiddleware.NewValidationMiddleware(cacheRepo))
	
	// Modificação do servidor para incluir timeout de shutdown
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", env.GetEnvAsInt(env.PORT, env.DEFAULT_PORT)),
		Handler: e,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal(ctx, err, "shutting down the server")
		}
	}()

	// Aguarda sinal de interrupção
	<-ctx.Done()

	// Cria um novo contexto com timeout para shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	// Tenta realizar shutdown gracefully
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Fatal(ctx, err, "erro durante o shutdown do servidor")
	}

	logger.Info(ctx, "servidor encerrado com sucesso")
}
