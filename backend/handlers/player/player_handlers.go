package players_handlers

import (
    internaldbgamestate "monopoly-backend/internal/db/game_state"
    internaldbplayers "monopoly-backend/internal/db/player"
    "monopoly-backend/util"
    "net/http"
    "strconv"
    "time"

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
        "id":         id,
        "name":       name,
        "session_id": sessionId,
    })
}

func JoinPlayerHandler(c echo.Context) error {
    log := util.GetRequestLogger(c)
    ctx := c.Request().Context()

    playerId := c.FormValue("player_id")
    if playerId == "" {
        return c.String(http.StatusBadRequest, "missing player_id")
    }

    name := c.FormValue("player_name")
    if name == "" {
        return c.String(http.StatusBadRequest, "missing player_name")
    }

    sessionId := c.FormValue("session_id")
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

    playerExists, err := internaldbplayers.CheckPlayerExists(log, ctx, tx, playerId, name, sessionId)
    if err != nil {
        return c.String(http.StatusInternalServerError, "failed to query db about player")
    }

    if !playerExists {
        return c.String(http.StatusBadRequest, "player does not exist")
    }

    playerIdInt, err := strconv.Atoi(playerId)
    if err != nil {
        return c.String(http.StatusBadRequest, "player_id must be a valid integer")
    }

    jwt, err := util.CreatePlayerJwt(playerIdInt, name, sessionId)
    if err != nil {
        return c.String(http.StatusInternalServerError, "failed to create player auth token")
    }

    c.SetCookie(&http.Cookie{
        Name:    util.PlayerAuthCookieName,
        Value:   jwt,
        Path:    "/",
        Secure:  true,
        Expires: time.Now().Add(24 * time.Hour),
    })

    return c.JSON(http.StatusOK, map[string]interface{}{
        "id":         playerIdInt,
        "name":       name,
        "session_id": sessionId,
        "jwt":        jwt,
    })
}
