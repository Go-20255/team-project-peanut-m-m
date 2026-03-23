package players_handlers

import (
    internaldbgamestate "monopoly-backend/internal/db/game_state"
    internaldbplayers "monopoly-backend/internal/db/players"
    "monopoly-backend/util"
    "net/http"

    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/labstack/echo/v4"
)

func CreatePlayerHandler(c echo.Context) error {
    log := util.GetRequestLogger(c)
    ctx := c.Request().Context()

    name := c.FormValue("player_name")
    if name == "" {
        return c.String(http.StatusBadRequest, "missing player name")
    }

    sessionId := c.FormValue("session_id")

    tx := c.Get("tx").(*pgxpool.Tx)
    err := internaldbgamestate.GameStateExists(log, ctx, tx, sessionId)
    if err != nil {
        return c.String(http.StatusBadRequest, "session_id does not exist")
    }

    err = internaldbplayers.CreatePlayerDB(log, ctx, tx, name, sessionId)
    if err != nil {
        return c.String(http.StatusInternalServerError, "failed to create player in db")
    }

    return c.String(http.StatusOK, "created player")
}

func JoinGameHandler(c echo.Context) error {
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

    return c.String(http.StatusOK, "joined game: "+id)
}
