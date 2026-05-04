package properties_handlers

import (
    "fmt"
    "net/http"
    "strconv"

    "monopoly-backend/internal"
    "monopoly-backend/internal/db/game_state"
    internaldb_tiles "monopoly-backend/internal/db/tile"
    "monopoly-backend/internal/engine"
    "monopoly-backend/util"

    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/labstack/echo/v4"
)

func CheckPropertyOwnerHandler(c echo.Context) error {
    log := util.GetRequestLogger(c)
    ctx := c.Request().Context()
    claims, err := util.GetPlayerJwtClaims(c)
    if err != nil {
        return c.String(http.StatusUnauthorized, err.Error())
    }

    propertyIdStr := c.QueryParam("id")

    if propertyIdStr == "" {
        return c.String(http.StatusBadRequest, "missing session_id or property_id")
    }

    propertyId, err := strconv.Atoi(propertyIdStr)
    if err != nil {
        return c.String(http.StatusBadRequest, "property_id must be a valid integer")
    }

    tx := c.Get("tx").(*pgxpool.Tx)
    exists, err := internaldb_game_state.GameStateExists(log, ctx, tx, claims.SessionId)
    if err != nil {
        return c.String(http.StatusInternalServerError, "failed to query db about game state")
    }

    if !exists {
        return c.String(http.StatusBadRequest, "session_id does not exist")
    }

    ownerId, owned, err := internaldb_tiles.VerifyPropertyOwnerDB(log, ctx, tx, claims.SessionId, propertyId)
    if err != nil {
        return c.String(http.StatusInternalServerError, "failed to get property ownership status")
    }

    return c.JSON(http.StatusOK, map[string]any{
        "owner_id":    ownerId,
        "owned":      owned,
    })
}

func PurchasePropertyHandler(c echo.Context) error {
    claims, err := util.GetPlayerJwtClaims(c)
    if err != nil {
        return c.String(http.StatusUnauthorized, err.Error())
    }

    res, err := monopoly_engine.NotifyEngineOfAction(claims.SessionId, internal.UserActionEvent{
        Event: "PurchaseProperty",
        Data: internal.SimpleActionData{
            SessionId:  claims.SessionId,
            PlayerId:   claims.PlayerId,
        },
        ReturnChan: make(chan internal.UserActionStatus),
    })

    if err != nil {
        return c.String(http.StatusInternalServerError, err.Error())
    }

    if res.Status != http.StatusOK {
        return c.String(res.Status, res.Msg)
    }

    return c.String(http.StatusOK, res.Msg)
}

func IgnorePropertyPurchaseHandler(c echo.Context) error {
    claims, err := util.GetPlayerJwtClaims(c)
    if err != nil {
        return c.String(http.StatusUnauthorized, err.Error())
    }

    res, err := monopoly_engine.NotifyEngineOfAction(claims.SessionId, internal.UserActionEvent{
        Event: "IgnorePropertyPurchase",
        Data: internal.SimpleActionData{
            SessionId: claims.SessionId,
            PlayerId:  claims.PlayerId,
        },
        ReturnChan: make(chan internal.UserActionStatus),
    })

    if err != nil {
        return c.String(http.StatusInternalServerError, err.Error())
    }

    if res.Status != http.StatusOK {
        return c.String(res.Status, res.Msg)
    }

    return c.String(http.StatusOK, res.Msg)
}

func PurchaseHouseHandler(c echo.Context) error {
    claims, err := util.GetPlayerJwtClaims(c)
    if err != nil {
        return c.String(http.StatusUnauthorized, err.Error())
    }

    propertyId, err := getPropertyIdFromForm(c)
    if err != nil {
        return c.String(http.StatusBadRequest, err.Error())
    }

    res, err := monopoly_engine.NotifyEngineOfAction(claims.SessionId, internal.UserActionEvent{
        Event: "PurchaseHouse",
        Data: internal.PropertyActionData{
            SessionId:  claims.SessionId,
            PlayerId:   claims.PlayerId,
            PropertyId: propertyId,
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

func PurchaseHotelHandler(c echo.Context) error {
    claims, err := util.GetPlayerJwtClaims(c)
    if err != nil {
        return c.String(http.StatusUnauthorized, err.Error())
    }

    propertyId, err := getPropertyIdFromForm(c)
    if err != nil {
        return c.String(http.StatusBadRequest, err.Error())
    }

    res, err := monopoly_engine.NotifyEngineOfAction(claims.SessionId, internal.UserActionEvent{
        Event: "PurchaseHotel",
        Data: internal.PropertyActionData{
            SessionId:  claims.SessionId,
            PlayerId:   claims.PlayerId,
            PropertyId: propertyId,
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

func SellHouseHandler(c echo.Context) error {
    claims, err := util.GetPlayerJwtClaims(c)
    if err != nil {
        return c.String(http.StatusUnauthorized, err.Error())
    }

    propertyId, err := getPropertyIdFromForm(c)
    if err != nil {
        return c.String(http.StatusBadRequest, err.Error())
    }

    res, err := monopoly_engine.NotifyEngineOfAction(claims.SessionId, internal.UserActionEvent{
        Event: "SellHouse",
        Data: internal.PropertyActionData{
            SessionId:  claims.SessionId,
            PlayerId:   claims.PlayerId,
            PropertyId: propertyId,
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

func SellHotelHandler(c echo.Context) error {
    claims, err := util.GetPlayerJwtClaims(c)
    if err != nil {
        return c.String(http.StatusUnauthorized, err.Error())
    }

    propertyId, err := getPropertyIdFromForm(c)
    if err != nil {
        return c.String(http.StatusBadRequest, err.Error())
    }

    res, err := monopoly_engine.NotifyEngineOfAction(claims.SessionId, internal.UserActionEvent{
        Event: "SellHotel",
        Data: internal.PropertyActionData{
            SessionId:  claims.SessionId,
            PlayerId:   claims.PlayerId,
            PropertyId: propertyId,
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

func MortgagePropertyHandler(c echo.Context) error {
    claims, err := util.GetPlayerJwtClaims(c)
    if err != nil {
        return c.String(http.StatusUnauthorized, err.Error())
    }

    propertyId, err := getPropertyIdFromForm(c)
    if err != nil {
        return c.String(http.StatusBadRequest, err.Error())
    }

    res, err := monopoly_engine.NotifyEngineOfAction(claims.SessionId, internal.UserActionEvent{
        Event: "MortgageProperty",
        Data: internal.PropertyActionData{
            SessionId:  claims.SessionId,
            PlayerId:   claims.PlayerId,
            PropertyId: propertyId,
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

func UnmortgagePropertyHandler(c echo.Context) error {
    claims, err := util.GetPlayerJwtClaims(c)
    if err != nil {
        return c.String(http.StatusUnauthorized, err.Error())
    }

    propertyId, err := getPropertyIdFromForm(c)
    if err != nil {
        return c.String(http.StatusBadRequest, err.Error())
    }

    res, err := monopoly_engine.NotifyEngineOfAction(claims.SessionId, internal.UserActionEvent{
        Event: "UnmortgageProperty",
        Data: internal.PropertyActionData{
            SessionId:  claims.SessionId,
            PlayerId:   claims.PlayerId,
            PropertyId: propertyId,
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

func getPropertyIdFromForm(c echo.Context) (int, error) {
    propertyIdStr := c.FormValue("property_id")
    if propertyIdStr == "" {
        return 0, fmt.Errorf("missing property_id")
    }

    propertyId, err := strconv.Atoi(propertyIdStr)
    if err != nil {
        return 0, fmt.Errorf("property_id must be a valid integer")
    }

    return propertyId, nil
}
