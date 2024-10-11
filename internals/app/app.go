package app

import (
	"context"
	"log"
	"log/slog"
	"loshon-api/internals/config"
	"loshon-api/internals/data"
	"os"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gorm.io/gorm"
)

type App struct {
	engine *echo.Echo
	db     *gorm.DB
	config *config.AppConfig
}

func NewApp() *App {
	config, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Fail to populate app config: %v", err)
	}
	e := echo.New()
	// create new application
	app := &App{
		engine: e,
		config: config,
	}

	clerk.SetKey(app.config.ClerkSecretKey)
	app.engine.Pre(middleware.RemoveTrailingSlash())
	app.engine.Use(middleware.RequestID())
	app.engine.Use(middleware.Recover())
	app.engine.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
	}))
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	app.engine.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:   true,
		LogURI:      true,
		LogError:    true,
		HandleError: true, // forwards error to the global error handler, so it can decide appropriate status code
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error == nil {
				logger.LogAttrs(context.Background(), slog.LevelInfo, "REQUEST",
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
				)
			} else {
				logger.LogAttrs(context.Background(), slog.LevelError, "REQUEST_ERROR",
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
					slog.String("err", v.Error.Error()),
				)
			}
			return nil
		},
	}))
	app.engine.Static("", "assets")

	db, err := data.OpenDB(app.config.PostgresUrl)
	if err != nil {
		log.Fatal(err)
	}
	app.db = db

	app.RegisterRoutes()

	return app
}

func (app *App) Run() error {
	addr := app.config.Addr
	if addr == "" {
		addr = ":80"
	}
	return app.engine.Start(addr)
}

func (app *App) RegisterRoutes() {
	api := app.engine.Group("/api")
	api.GET("", app.healthCheck)
	api.GET("/document", app.GetDocuments, app.ClerkAuthMiddleware)
	api.POST("/document", app.CreateDocument, app.ClerkAuthMiddleware)
}

func (app *App) RunMigrate() {
	app.db.AutoMigrate(data.Document{})
}
