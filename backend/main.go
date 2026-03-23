package main

import (
	"context"
	common_handler "monopoly-backend/handlers/common"
	players_handlers "monopoly-backend/handlers/player"
	internaldb "monopoly-backend/internal/db"
	"monopoly-backend/util"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"
)



func main() {
    // create context to be used for lifetime of server
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    util.Setup_logging()
    log := log.Logger.With().Logger()

    log.Info().Msg("Welcome To Monopoly")
    err := godotenv.Load("../.internal.env")
    if err != nil {
        log.Info().Msg("WARNING: Failed to load .internal.env file in repo root")
    }

    
    internaldb.SetupDatabase(ctx, log)

    db, err := internaldb.CreateDbPoolConnection(ctx, log)
    if err != nil {
        log.Panic().Err(err).Msg("failed to connect to database")
        return
    }

    e := echo.New()

    // allow us to use a custom logger for each api call
    e.Use(util.RequestLoggerMiddleware)
    // attach a database transaction to each api call
    e.Use(internaldb.TxMiddleware(db))

    // setup cors
    e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
        AllowOrigins:     []string{
            "http://localhost:3001",
            "http://localhost:3000",
            "http://127.0.0.1:3001",
            "http://127.0.0.1:3000",
        },
        AllowMethods:     []string{"GET", "POST", "DELETE", "PUT", "PATCH"},
        AllowHeaders:     []string{"Authorization", "Content-Type"},
        AllowCredentials: true,
        ExposeHeaders:    []string{"Sec-WebSocket-Accept", "Sec-WebSocket-Protocol"},
    }))
    e.Use(middleware.Gzip())

    routes := e.Group("/api")

    // add routes here
    routes.GET("/health", common_handler.HealthCheckHandler)

    routes.POST("/player", players_handlers.CreatePlayerHandler)

    // start the echo server
    e.Start(":9876")
}
