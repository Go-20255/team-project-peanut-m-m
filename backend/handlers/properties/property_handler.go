package properties_handlers

import (
    "net/http"
    "strconv"

    "monopoly-backend/internal"
	"monopoly-backend/util"
    monopolyengine "monopoly-backend/internal/engine"
	internaldbgamestate "monopoly-backend/internal/db/game_state"
	internaldbproperties "monopoly-backend/internal/db/properties"

	"github.com/jackc/pgx/v5/pgxpool"
    "github.com/labstack/echo/v4"
)

func CheckPropertyOwnerHandler(c echo.Context) error {
    log := util.GetRequestLogger(c)
    ctx := c.Request().Context()

	sessionId := c.QueryParam("session_id")
    propertyIdStr := c.QueryParam("property_id")

    if sessionId == "" || propertyIdStr == "" {
        return c.String(http.StatusBadRequest, "missing session_id or property_id")
    }

    propertyId, err := strconv.Atoi(propertyIdStr)
    if err != nil {
        return c.String(http.StatusBadRequest, "property_id must be a valid integer")
    }

	tx := c.Get("tx").(*pgxpool.Tx)
    exists, err := internaldbgamestate.GameStateExists(log, ctx, tx, sessionId)
    if err != nil {
        return c.String(http.StatusInternalServerError, "failed to query db about game state")
    }
    
    if !exists {
        return c.String(http.StatusBadRequest, "session_id does not exist")
    }

	ownerId, owned, err := internaldbproperties.VerifyPropertyOwnerDB(log, ctx, tx, sessionId, propertyId)
	if err != nil {
		return c.String(http.StatusInternalServerError, "failed to get property ownership status")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
        "ownerId":      ownerId,
        "owned":        owned,
        "session_id":   sessionId,
    })
}

func PurchasePropertyHandler(c echo.Context) error {
    sessionId := c.FormValue("session_id")
    playerIdStr := c.FormValue("player_id")
    propertyIdStr := c.FormValue("property_id")

    if sessionId == "" || playerIdStr == "" || propertyIdStr == "" {
        return c.String(http.StatusBadRequest, "missing session_id, player_id, or property_id")
    }

    playerId, err := strconv.Atoi(playerIdStr)
    if err != nil {
        return c.String(http.StatusBadRequest, "player_id must be a valid integer")
    }

    propertyId, err := strconv.Atoi(propertyIdStr)
    if err != nil {
        return c.String(http.StatusBadRequest, "property_id must be a valid integer")
    }

    res, err := monopolyengine.NotifyEngineOfAction(sessionId, internal.UserActionEvent{
        Event: "PurchaseProperty",
        Data: struct {
            SessionId  string
            PlayerId   int
            PropertyId int
        }{
            SessionId:  sessionId,
            PlayerId:   playerId,
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

    return c.String(http.StatusOK, res.Msg)
}
