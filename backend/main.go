package main

import (
    "context"
    commonhandler "monopoly-backend/handlers/common"
    gamestatehandlers "monopoly-backend/handlers/game_state"
    playershandlers "monopoly-backend/handlers/player"
    properties_handlers "monopoly-backend/handlers/property"
    internaldb "monopoly-backend/internal/db"
    internaldbgamestate "monopoly-backend/internal/db/game_state"
    monopolyengine "monopoly-backend/internal/engine"
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

    util.SetupLogging()
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

    tempTx, err := db.BeginTx(ctx, pgx.TxOptions{})
    if err != nil {
        log.Fatal().Err(err).Msg("failed to create temporary transaction")
    }
    var sessionIds []string
    sessionIds, err = internaldbgamestate.GetGameSessions(log, ctx, tempTx.(*pgxpool.Tx))
    if err != nil {
        _ = tempTx.Rollback(ctx)
        log.Fatal().Msg("failed to query game sessions.")
    }
    if err := tempTx.Commit(ctx); err != nil {
        log.Fatal().Msg("failed to commit transaction")
    }

    // start up new monopoly engine for each found session_id
    for _, sessionId := range sessionIds {
        go monopolyengine.SetupNewMonopolyEngine(sessionId)
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
    routes.PATCH("/player", playershandlers.UpdatePlayerTokenHandler)
    routes.GET("/game/players", playershandlers.GetPlayersHandler)
    routes.POST("/player/join", playershandlers.JoinPlayerHandler)
    routes.POST("/player/readyup", playershandlers.ReadyUpPlayerHandler)
    routes.POST("/player/endturn", playershandlers.EndTurnHandler)

    routes.POST("/game", gamestatehandlers.NewGameHandler)
    routes.GET("/game", gamestatehandlers.GetAllGameSessions)
    routes.POST("/game/join", gamestatehandlers.JoinSessionHandler)
    routes.GET("/game/join/live", gamestatehandlers.JoinLiveGameHandler)
    routes.POST("/game/roll", gamestatehandlers.RollDiceHandler)
    routes.POST("/game/move", gamestatehandlers.MovePlayerHandler)
    routes.POST("/game/jail/release", gamestatehandlers.UseGetOutOfJailCardHandler)
    routes.POST("/game/bank/pay", gamestatehandlers.PayBankHandler)
    routes.POST("/game/bank/get", gamestatehandlers.SetBankPayoutHandler)
    routes.POST("/game/bank/receive", gamestatehandlers.ReceiveBankPayoutHandler)
    routes.POST("/game/bankrupt", gamestatehandlers.BankruptHandler)
    routes.POST("/game/rent", gamestatehandlers.PayRentHandler)

    routes.GET("/game/property", properties_handlers.CheckPropertyOwnerHandler)
    routes.POST("/game/property", properties_handlers.PurchasePropertyHandler)
    routes.POST("/game/property/house", properties_handlers.PurchaseHouseHandler)
    routes.POST("/game/property/hotel", properties_handlers.PurchaseHotelHandler)
    routes.POST("/game/property/house/sell", properties_handlers.SellHouseHandler)
    routes.POST("/game/property/hotel/sell", properties_handlers.SellHotelHandler)
    routes.POST("/game/property/mortgage", properties_handlers.MortgagePropertyHandler)
    routes.POST("/game/property/unmortgage", properties_handlers.UnmortgagePropertyHandler)

    // start the echo server
    e.Start(":9876")
}
