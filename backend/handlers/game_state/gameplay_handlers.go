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

func PayRentHandler(c echo.Context) error {
    sessionId := c.FormValue("session_id")
    if sessionId == "" {
        return c.String(http.StatusBadRequest, "missing session_id")
    }

    fromPlayerIdStr := c.FormValue("from_player_id")
    if fromPlayerIdStr == "" {
        return c.String(http.StatusBadRequest, "missing from_player_id")
    }

    toPlayerIdStr := c.FormValue("to_player_id")
    if toPlayerIdStr == "" {
        return c.String(http.StatusBadRequest, "missing to_player_id")
    }

    amountStr := c.FormValue("amount")
    if amountStr == "" {
        return c.String(http.StatusBadRequest, "missing amount")
    }

    fromPlayerId, err := strconv.Atoi(fromPlayerIdStr)
    if err != nil {
        return c.String(http.StatusBadRequest, "invalid from_player_id")
    }

    toPlayerId, err := strconv.Atoi(toPlayerIdStr)
    if err != nil {
        return c.String(http.StatusBadRequest, "invalid to_player_id")
    }

    amount, err := strconv.Atoi(amountStr)
    if err != nil {
        return c.String(http.StatusBadRequest, "invalid amount")
    }

    res, err := monopolyengine.NotifyEngineOfAction(sessionId, internal.UserActionEvent{
        Event: "PayRentEvent",
        Data: internal.RentPaymentActionData{
            FromPlayerId: fromPlayerId,
            ToPlayerId:   toPlayerId,
            SessionId:    sessionId,
            Amount:       amount,
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
