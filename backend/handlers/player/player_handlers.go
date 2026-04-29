package players_handlers

import (
	internaldbgamestate "monopoly-backend/internal/db/game_state"
    internaldbplayers "monopoly-backend/internal/db/player"
    "monopoly-backend/util"
    "net/http"
    "strconv"

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

    pieceToken, err := util.AssignPlayerToken(log, ctx, tx, sessionId)
    if err != nil {
        return c.String(http.StatusInternalServerError, "failed to assign player token")
    }

    id, err := internaldbplayers.CreatePlayerDB(log, ctx, tx, name, sessionId, pieceToken)
    if err != nil {
        return c.String(http.StatusInternalServerError, "failed to create player in db")
    }

    return c.JSON(http.StatusOK, map[string]interface{}{
        "id":           id,
        "name":         name,
        "session_id":   sessionId,
        "piece_token":  pieceToken,
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

func UpdatePlayerTokenHandler(c echo.Context) error {
    log := util.GetRequestLogger(c)
    ctx := c.Request().Context()

    playerId, _ := strconv.Atoi(c.FormValue("player_id"))
    sessionId := c.FormValue("session_id")
    pieceToken, _ := strconv.Atoi(c.FormValue("piece_token"))

    tx := c.Get("tx").(*pgxpool.Tx)
    err := internaldbplayers.UpdatePlayerPieceToken(log, ctx, tx, playerId, sessionId, pieceToken)
    if err != nil {
        return c.String(http.StatusInternalServerError, "failed to update player token")
    }

    return c.JSON(http.StatusOK, map[string]interface{}{
        "id":           playerId,
        "session_id":   sessionId,
        "piece_token":  pieceToken,
    })
}
