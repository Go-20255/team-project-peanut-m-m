package main

import (
	"context"
	commonhandler "monopoly-backend/handlers/common"
	gamestatehandlers "monopoly-backend/handlers/game_state"
	playershandlers "monopoly-backend/handlers/player"
	internaldb "monopoly-backend/internal/db"
	internaldb_game_state "monopoly-backend/internal/db/game_state"
	monopoly_engine "monopoly-backend/internal/engine"
	"monopoly-backend/util"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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

    temp_tx, err := db.BeginTx(ctx, pgx.TxOptions{})
    if err != nil {
        log.Fatal().Err(err).Msg("failed to create temporary transaction")
    }
    var session_ids []string
    session_ids, err = internaldb_game_state.GetGameSessions(log, ctx, temp_tx.(*pgxpool.Tx))
    if err != nil {
        _ = temp_tx.Rollback(ctx)
        log.Fatal().Msg("failed to query game sessions.")
    }
    if err := temp_tx.Commit(ctx); err != nil {
        log.Fatal().Msg("failed to commit transaction")
    }

    // start up new monopoly engine for each found session_id
    for _, session_id := range session_ids {
        go monopoly_engine.SetupNewMonopolyEngine(session_id)
    }


    e := echo.New()

    // allow us to use a custom logger for each api call
    e.Use(util.RequestLoggerMiddleware)
    // attach a database transaction to each api call
    e.Use(internaldb.TxMiddleware(db))

    // setup cors
    e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
        AllowOrigins: []string{
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
    routes.GET("/health", commonhandler.HealthCheckHandler)

    routes.POST("/player", playershandlers.CreatePlayerHandler)

    routes.POST("/game", gamestatehandlers.NewGameHandler)
    routes.POST("/game/join", gamestatehandlers.JoinSessionHandler)
    routes.GET("/game/join/live", gamestatehandlers.JoinLiveGameHandler)

    // start the echo server
    e.Start(":9876")
}
