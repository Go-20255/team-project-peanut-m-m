package game_state_handlers

import (
    "monopoly-backend/internal"
    monopolyengine "monopoly-backend/internal/engine"
    "monopoly-backend/util"
    "net/http"
    "strconv"
    "strings"

    "github.com/labstack/echo/v4"
)

func parseTradePropertyIds(c echo.Context, key string) ([]int, error) {
    values := c.Request().Form[key]
    if len(values) == 0 {
        return []int{}, nil
    }

    propertyIds := make([]int, 0)
    for _, value := range values {
        if value == "" {
            continue
        }

        for _, part := range strings.Split(value, ",") {
            trimmed := strings.TrimSpace(part)
            if trimmed == "" {
                continue
            }

            propertyId, err := strconv.Atoi(trimmed)
            if err != nil {
                return nil, err
            }

            propertyIds = append(propertyIds, propertyId)
        }
    }

    return propertyIds, nil
}

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

func DrawCardHandler(c echo.Context) error {
    claims, err := util.GetPlayerJwtClaims(c)
    if err != nil {
        return c.String(http.StatusUnauthorized, err.Error())
    }

    res, err := monopolyengine.NotifyEngineOfAction(claims.SessionId, internal.UserActionEvent{
        Event: "DrawCardEvent",
        Data: internal.CardActionData{
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

func ResolveCardHandler(c echo.Context) error {
    claims, err := util.GetPlayerJwtClaims(c)
    if err != nil {
        return c.String(http.StatusUnauthorized, err.Error())
    }

    res, err := monopolyengine.NotifyEngineOfAction(claims.SessionId, internal.UserActionEvent{
        Event: "ResolveCardEvent",
        Data: internal.CardActionData{
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

    if res.Data == nil {
        return c.String(http.StatusOK, res.Msg)
    }

    return c.JSON(http.StatusOK, res.Data)
}

func PayRentHandler(c echo.Context) error {
    claims, err := util.GetPlayerJwtClaims(c)
    if err != nil {
        return c.String(http.StatusUnauthorized, err.Error())
    }

    toPlayerIdStr := c.QueryParam("to_player_id")
    if toPlayerIdStr == "" {
        return c.String(http.StatusBadRequest, "missing to_player_id")
    }

    amountStr := c.QueryParam("amount")
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

    amountStr := c.QueryParam("amount")
    if amountStr == "" {
        return c.String(http.StatusBadRequest, "missing amount")
    }

    amount, err := strconv.Atoi(amountStr)
    if err != nil {
        return c.String(http.StatusBadRequest, "invalid amount")
    }

    reason := c.QueryParam("reason")
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

func ExecutePlayerExchangeHandler(c echo.Context) error {
    claims, err := util.GetPlayerJwtClaims(c)
    if err != nil {
        return c.String(http.StatusUnauthorized, err.Error())
    }

    res, err := monopolyengine.NotifyEngineOfAction(claims.SessionId, internal.UserActionEvent{
        Event: "PlayerExchangeEvent",
        Data: internal.PlayerExchangeActionData{
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

func OpenTradeDraftHandler(c echo.Context) error {
    claims, err := util.GetPlayerJwtClaims(c)
    if err != nil {
        return c.String(http.StatusUnauthorized, err.Error())
    }

    if err := c.Request().ParseForm(); err != nil {
        return c.String(http.StatusBadRequest, "failed to parse trade request")
    }

    withPlayerIdStr := c.FormValue("with_player_id")
    if withPlayerIdStr == "" {
        return c.String(http.StatusBadRequest, "missing with_player_id")
    }

    withPlayerId, err := strconv.Atoi(withPlayerIdStr)
    if err != nil {
        return c.String(http.StatusBadRequest, "invalid with_player_id")
    }

    res, err := monopolyengine.NotifyEngineOfAction(claims.SessionId, internal.UserActionEvent{
        Event: "OpenTradeDraftEvent",
        Data: internal.TradeDraftActionData{
            PlayerId:     claims.PlayerId,
            SessionId:    claims.SessionId,
            WithPlayerId: withPlayerId,
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

func CloseTradeDraftHandler(c echo.Context) error {
    claims, err := util.GetPlayerJwtClaims(c)
    if err != nil {
        return c.String(http.StatusUnauthorized, err.Error())
    }

    res, err := monopolyengine.NotifyEngineOfAction(claims.SessionId, internal.UserActionEvent{
        Event: "CloseTradeDraftEvent",
        Data: internal.TradeDecisionActionData{
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

func ProposeTradeHandler(c echo.Context) error {
    claims, err := util.GetPlayerJwtClaims(c)
    if err != nil {
        return c.String(http.StatusUnauthorized, err.Error())
    }

    if err := c.Request().ParseForm(); err != nil {
        return c.String(http.StatusBadRequest, "failed to parse trade request")
    }

    withPlayerIdStr := c.FormValue("with_player_id")
    if withPlayerIdStr == "" {
        return c.String(http.StatusBadRequest, "missing with_player_id")
    }

    withPlayerId, err := strconv.Atoi(withPlayerIdStr)
    if err != nil {
        return c.String(http.StatusBadRequest, "invalid with_player_id")
    }

    offeredMoney := 0
    offeredMoneyStr := c.FormValue("offered_money")
    if offeredMoneyStr != "" {
        offeredMoney, err = strconv.Atoi(offeredMoneyStr)
        if err != nil {
            return c.String(http.StatusBadRequest, "invalid offered_money")
        }
    }

    requestedMoney := 0
    requestedMoneyStr := c.FormValue("requested_money")
    if requestedMoneyStr != "" {
        requestedMoney, err = strconv.Atoi(requestedMoneyStr)
        if err != nil {
            return c.String(http.StatusBadRequest, "invalid requested_money")
        }
    }

    offeredPropertyIds, err := parseTradePropertyIds(c, "offered_property_ids")
    if err != nil {
        return c.String(http.StatusBadRequest, "invalid offered_property_ids")
    }

    requestedPropertyIds, err := parseTradePropertyIds(c, "requested_property_ids")
    if err != nil {
        return c.String(http.StatusBadRequest, "invalid requested_property_ids")
    }

    res, err := monopolyengine.NotifyEngineOfAction(claims.SessionId, internal.UserActionEvent{
        Event: "ProposeTradeEvent",
        Data: internal.TradeActionData{
            PlayerId:             claims.PlayerId,
            SessionId:            claims.SessionId,
            WithPlayerId:         withPlayerId,
            OfferedMoney:         offeredMoney,
            RequestedMoney:       requestedMoney,
            OfferedPropertyIds:   offeredPropertyIds,
            RequestedPropertyIds: requestedPropertyIds,
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

func AcceptTradeHandler(c echo.Context) error {
    claims, err := util.GetPlayerJwtClaims(c)
    if err != nil {
        return c.String(http.StatusUnauthorized, err.Error())
    }

    res, err := monopolyengine.NotifyEngineOfAction(claims.SessionId, internal.UserActionEvent{
        Event: "AcceptTradeEvent",
        Data: internal.TradeDecisionActionData{
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

func RejectTradeHandler(c echo.Context) error {
    claims, err := util.GetPlayerJwtClaims(c)
    if err != nil {
        return c.String(http.StatusUnauthorized, err.Error())
    }

    res, err := monopolyengine.NotifyEngineOfAction(claims.SessionId, internal.UserActionEvent{
        Event: "RejectTradeEvent",
        Data: internal.TradeDecisionActionData{
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

func CancelTradeHandler(c echo.Context) error {
    claims, err := util.GetPlayerJwtClaims(c)
    if err != nil {
        return c.String(http.StatusUnauthorized, err.Error())
    }

    res, err := monopolyengine.NotifyEngineOfAction(claims.SessionId, internal.UserActionEvent{
        Event: "CancelTradeEvent",
        Data: internal.TradeDecisionActionData{
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

func BankruptHandler(c echo.Context) error {
    claims, err := util.GetPlayerJwtClaims(c)
    if err != nil {
        return c.String(http.StatusUnauthorized, err.Error())
    }

    res, err := monopolyengine.NotifyEngineOfAction(claims.SessionId, internal.UserActionEvent{
        Event: "BankruptEvent",
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

func UseGetOutOfJailCardHandler(c echo.Context) error {
    claims, err := util.GetPlayerJwtClaims(c)
    if err != nil {
        return c.String(http.StatusUnauthorized, err.Error())
    }

    method := c.QueryParam("method")
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
