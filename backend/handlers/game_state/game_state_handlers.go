package game_state_handlers

import (
    internaldbgamestate "monopoly-backend/internal/db/game_state"
    monopolyengine "monopoly-backend/internal/engine"
    "monopoly-backend/util"
    "net/http"

    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/labstack/echo/v4"
)

func JoinSessionHandler(c echo.Context) error {
    log := util.GetRequestLogger(c)
    ctx := c.Request().Context()

    code := c.QueryParam("code")
    if code == "" {
        return c.String(http.StatusBadRequest, "missing code")
    }

    tx := c.Get("tx").(*pgxpool.Tx)
    id, err := internaldbgamestate.CheckGameStateCode(log, ctx, tx, code)
    if err != nil {
        return c.String(http.StatusBadRequest, "code does not exist")
    }

    return c.String(http.StatusOK, id)
}

func NewGameHandler(c echo.Context) error {
    log := util.GetRequestLogger(c)
    ctx := c.Request().Context()

    tx := c.Get("tx").(*pgxpool.Tx)
    sessionId, err := internaldbgamestate.CreateGameState(log, ctx, tx)
    if err != nil {
        return c.String(http.StatusInternalServerError, "failed to create new monopoly game")
    }

    code, err := internaldbgamestate.GetGameStateCode(log, ctx, tx, sessionId)
    if err != nil {
        return c.String(http.StatusInternalServerError, "failed to retrieve game code")
    }

    go monopolyengine.SetupNewMonopolyEngine(sessionId)

    return c.JSON(http.StatusOK, map[string]interface{}{
        "session_id": sessionId,
        "code":       code,
    })
}

func GetAllGameSessions(c echo.Context) error {
    log := util.GetRequestLogger(c)
    ctx := c.Request().Context()

    tx := c.Get("tx").(*pgxpool.Tx)
    sessionIds, err := internaldbgamestate.GetGameSessions(log, ctx, tx)
    if err != nil {
        return c.String(http.StatusInternalServerError, "failed to get list of game sessions")
    }

    return c.JSON(http.StatusOK, sessionIds)
}


