package app

import (
	"context"
	"log"
	"log/slog"
	"loshon-api/internals/config"
	"loshon-api/internals/data"
	"loshon-api/internals/search"
	"os"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type App struct {
	engine  *echo.Echo
	config  *config.AppConfig
	sclient *search.SearchClient
    documentRepo data.DocumentRepositoryInterface
}

func NewApp() *App {
	e := echo.New()
	// create new application
	app := &App{
		engine: e,
	}

	app.RegisterConfig()
	clerk.SetKey(app.config.ClerkSecretKey)
	app.RegisterMiddlewares()
	app.RegisterRepos()
	app.RegisterSearchClient()
	app.RegisterRoutes()

	return app
}

func (app *App) RegisterConfig() {
	var err error
	app.config, err = config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load app config %v", err)
	}
}

func (app *App) RegisterRepos() {
	db, err := data.OpenDB(app.config.PostgresUrl)
	if err != nil {
		log.Fatal(err)
	}
	app.documentRepo = data.NewDocumentRepository(db)
}

func (app *App) RegisterSearchClient() {
	sclient, err := search.NewSearchClient(app.config.AngoliaAppID, app.config.AngoliaAPIKey)
	if err != nil {
		log.Fatalf("cannot initialize search client %v", err)
	}
	app.sclient = sclient
}

func (app *App) RegisterMiddlewares() {
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
}

func (app *App) RegisterRoutes() {
	api := app.engine.Group("/api")

	api.GET("", app.healthCheck)

	api.GET("/documents", app.GetDocuments, app.ClerkAuthMiddleware)
	api.GET("/documents/:documentID", app.GetDocumentByID, app.OptionalClerkAuthMiddleware)
	api.POST("/documents", app.CreateDocument, app.ClerkAuthMiddleware)
	api.PATCH("/documents/:documentID", app.UpdateDocument, app.ClerkAuthMiddleware)
	api.DELETE("/documents/:documentID", app.ArchiveDocument, app.ClerkAuthMiddleware)

	api.GET("/documents/_archives", app.GetArchivedDocuments, app.ClerkAuthMiddleware)
	api.PATCH("/documents/_restore/:documentID", app.RestoreArchivedDocument, app.ClerkAuthMiddleware)
	api.DELETE("/documents/_delete/:documentID", app.DeleteArchivedDocument, app.ClerkAuthMiddleware)
}

func (app *App) Run() error {
	addr := app.config.Port
	if addr == "" {
		addr = ":80"
	}
	return app.engine.Start(":" + addr)
}

func (a App) healthCheck(c echo.Context) error {
	return c.JSON(200, map[string]string{
		"ping": "pong",
	})
}
