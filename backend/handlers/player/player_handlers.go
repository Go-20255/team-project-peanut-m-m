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
        return c.String(http.StatusBadRequest, "missing player_name")
    }

    sessionId := c.FormValue("session_id")
	if name == "" {
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

    id, err := internaldbplayers.CreatePlayerDB(log, ctx, tx, name, sessionId)
    if err != nil {
        return c.String(http.StatusInternalServerError, "failed to create player in db")
    }

    return c.JSON(http.StatusOK, map[string]interface{}{
        "id":           id,
        "name":         name,
        "session_id":   sessionId,
    })
}
