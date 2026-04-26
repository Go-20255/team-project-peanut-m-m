package players_handlers

import (
    internaldbgamestate "monopoly-backend/internal/db/game_state"
    internaldbplayers "monopoly-backend/internal/db/player"
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
        return c.String(http.StatusBadRequest, "missing player_name")
    }

    sessionId := c.FormValue("session_id")
	if sessionId == "" {
    if sessionId == "" {
        return c.String(http.StatusBadRequest, "missing session_id")
    }

    tx := c.Get("tx").(*pgxpool.Tx)
    exists, err := internaldbgamestate.GameStateExists(log, ctx, tx, sessionId)
    if err != nil {
        return c.String(http.StatusInternalServerError, "failed to query db about game state")
    }

    if !exists {
        return c.String(http.StatusBadRequest, "session_id does not exist")
    }

    // assigns a color to a player - TODO: need to remove this and use tokens instead
    color, err := util.AssignPlayerColor(log, ctx, tx, sessionId)
    if err != nil {
        return c.String(http.StatusInternalServerError, "failed to assign player color")
    }

    id, err := internaldbplayers.CreatePlayerDB(log, ctx, tx, name, sessionId, color)
    if err != nil {
        return c.String(http.StatusInternalServerError, "failed to create player in db")
    }

    return c.JSON(http.StatusOK, map[string]interface{}{
        "id":         id,
        "name":       name,
        "session_id": sessionId,
    })
}

func GetPlayersHandler(c echo.Context) error {
    log := util.GetRequestLogger(c)
    ctx := c.Request().Context()

    sessionId := c.QueryParam("session_id")
    if sessionId == "" {
        return c.String(http.StatusBadRequest, "missing session_id")
    }

    tx := c.Get("tx").(*pgxpool.Tx)
    
    players, err := internaldbplayers.GetPlayersInSession(log, ctx, tx, sessionId)
    if err != nil {
        return c.String(http.StatusInternalServerError, "failed to fetch players")
    }

    return c.JSON(http.StatusOK, players)
}
