package game_state_handlers

import (
    "monopoly-backend/internal"
    monopolyengine "monopoly-backend/internal/engine"
    "monopoly-backend/util"
    "net/http"
    "strconv"

    "github.com/labstack/echo/v4"
)

func RollDiceHandler(c echo.Context) error {
    claims, err := util.GetPlayerJwtClaims(c)
    if err != nil {
        return c.String(http.StatusUnauthorized, err.Error())
    }

    res, err := monopolyengine.NotifyEngineOfAction(claims.SessionId, internal.UserActionEvent{
        Event: "RollDiceEvent",
        Data: internal.SimpleActionData{
            PlayerId:  claims.PlayerId,
            SessionId: claims.SessionId,
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
    claims, err := util.GetPlayerJwtClaims(c)
    if err != nil {
        return c.String(http.StatusUnauthorized, err.Error())
    }

    res, err := monopolyengine.NotifyEngineOfAction(claims.SessionId, internal.UserActionEvent{
        Event: "MovePlayerEvent",
        Data: internal.SimpleActionData{
            PlayerId:  claims.PlayerId,
            SessionId: claims.SessionId,
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
    claims, err := util.GetPlayerJwtClaims(c)
    if err != nil {
        return c.String(http.StatusUnauthorized, err.Error())
    }

    toPlayerIdStr := c.FormValue("to_player_id")
    if toPlayerIdStr == "" {
        return c.String(http.StatusBadRequest, "missing to_player_id")
    }

    amountStr := c.FormValue("amount")
    if amountStr == "" {
        return c.String(http.StatusBadRequest, "missing amount")
    }

    toPlayerId, err := strconv.Atoi(toPlayerIdStr)
    if err != nil {
        return c.String(http.StatusBadRequest, "invalid to_player_id")
    }

    amount, err := strconv.Atoi(amountStr)
    if err != nil {
        return c.String(http.StatusBadRequest, "invalid amount")
    }

    res, err := monopolyengine.NotifyEngineOfAction(claims.SessionId, internal.UserActionEvent{
        Event: "PayRentEvent",
        Data: internal.RentPaymentActionData{
            FromPlayerId: claims.PlayerId,
            ToPlayerId:   toPlayerId,
            SessionId:    claims.SessionId,
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

func PayBankHandler(c echo.Context) error {
    claims, err := util.GetPlayerJwtClaims(c)
    if err != nil {
        return c.String(http.StatusUnauthorized, err.Error())
    }

    res, err := monopolyengine.NotifyEngineOfAction(claims.SessionId, internal.UserActionEvent{
        Event: "BankPaymentEvent",
        Data: internal.BankPaymentActionData{
            PlayerId:  claims.PlayerId,
            SessionId: claims.SessionId,
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

func SetBankPayoutHandler(c echo.Context) error {
    claims, err := util.GetPlayerJwtClaims(c)
    if err != nil {
        return c.String(http.StatusUnauthorized, err.Error())
    }

    amountStr := c.FormValue("amount")
    if amountStr == "" {
        return c.String(http.StatusBadRequest, "missing amount")
    }

    amount, err := strconv.Atoi(amountStr)
    if err != nil {
        return c.String(http.StatusBadRequest, "invalid amount")
    }

    reason := c.FormValue("reason")
    if reason == "" {
        reason = "bank payout"
    }

    res, err := monopolyengine.NotifyEngineOfAction(claims.SessionId, internal.UserActionEvent{
        Event: "SetBankPayoutEvent",
        Data: internal.BankPayoutActionData{
            PlayerId:  claims.PlayerId,
            SessionId: claims.SessionId,
            Amount:    amount,
            Reason:    reason,
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

func ReceiveBankPayoutHandler(c echo.Context) error {
    claims, err := util.GetPlayerJwtClaims(c)
    if err != nil {
        return c.String(http.StatusUnauthorized, err.Error())
    }

    res, err := monopolyengine.NotifyEngineOfAction(claims.SessionId, internal.UserActionEvent{
        Event: "BankPayoutEvent",
        Data: internal.BankPaymentActionData{
            PlayerId:  claims.PlayerId,
            SessionId: claims.SessionId,
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

func UseGetOutOfJailCardHandler(c echo.Context) error {
    claims, err := util.GetPlayerJwtClaims(c)
    if err != nil {
        return c.String(http.StatusUnauthorized, err.Error())
    }

    method := c.FormValue("method")
    if method == "" {
        method = "card"
    }

    if method != "card" && method != "pay" {
        return c.String(http.StatusBadRequest, "invalid method")
    }

    res, err := monopolyengine.NotifyEngineOfAction(claims.SessionId, internal.UserActionEvent{
        Event: "ReleaseFromJailEvent",
        Data: internal.JailReleaseActionData{
            PlayerId:  claims.PlayerId,
            SessionId: claims.SessionId,
            Method:    method,
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
