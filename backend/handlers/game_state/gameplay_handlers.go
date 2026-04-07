package game_state_handlers

import (
    "monopoly-backend/internal"
    monopolyengine "monopoly-backend/internal/engine"
    "net/http"
    "strconv"

    "github.com/labstack/echo/v4"
)

func RollDiceHandler(c echo.Context) error {
    sessionId := c.FormValue("session_id")
    if sessionId == "" {
        return c.String(http.StatusBadRequest, "missing session_id")
    }

    playerIdStr := c.FormValue("player_id")
    if playerIdStr == "" {
        return c.String(http.StatusBadRequest, "missing player_id")
    }

    playerId, err := strconv.Atoi(playerIdStr)
    if err != nil {
        return c.String(http.StatusBadRequest, "invalid player_id")
    }

    res, err := monopolyengine.NotifyEngineOfAction(sessionId, internal.UserActionEvent{
        Event: "RollDiceEvent",
        Data: internal.RollDiceActionData{
            PlayerId:  playerId,
            SessionId: sessionId,
        },
        ReturnChan: make(chan internal.UserActionStatus),
    })
    if err != nil {
        return c.String(http.StatusInternalServerError, err.Error())
    }

    if res.Status != http.StatusOK {
        return c.String(res.Status, res.Msg)
    }

    return c.JSON(http.StatusOK, res.Data)
}

func MovePlayerHandler(c echo.Context) error {
    sessionId := c.FormValue("session_id")
    if sessionId == "" {
        return c.String(http.StatusBadRequest, "missing session_id")
    }

    playerIdStr := c.FormValue("player_id")
    if playerIdStr == "" {
        return c.String(http.StatusBadRequest, "missing player_id")
    }

    playerId, err := strconv.Atoi(playerIdStr)
    if err != nil {
        return c.String(http.StatusBadRequest, "invalid player_id")
    }

    res, err := monopolyengine.NotifyEngineOfAction(sessionId, internal.UserActionEvent{
        Event: "MovePlayerEvent",
        Data: internal.MovePlayerActionData{
            PlayerId:  playerId,
            SessionId: sessionId,
        },
        ReturnChan: make(chan internal.UserActionStatus),
    })
    if err != nil {
        return c.String(http.StatusInternalServerError, err.Error())
    }

    if res.Status != http.StatusOK {
        return c.String(res.Status, res.Msg)
    }

    return c.JSON(http.StatusOK, res.Data)
}
