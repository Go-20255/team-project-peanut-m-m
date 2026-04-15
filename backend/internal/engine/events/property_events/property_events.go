package property_events

import (
	"context"
	"fmt"
	"monopoly-backend/internal"
	internaldb_properties "monopoly-backend/internal/db/properties"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

func PurchaseProperty (
    ctx context.Context,
    log zerolog.Logger,
    e *internal.MonopolyEngine,
    action *internal.UserActionEvent,
    tx *pgxpool.Tx,
) (internal.UserActionStatus) {
    log.Trace().Msg("player attempting to purchase property")

    data, ok := action.Data.(struct {
        SessionId  string
        PlayerId   int
        PropertyId int
    })
    if !ok {
        log.Error().Msg("invalid data format received for PurchaseProperty")
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "internal data format error",
        }
    }

    // try to buy property, also does ownership validation
    ownershipId, err := internaldb_properties.CreatePropertyOwnerDB(
        log,
        ctx,
        tx,
        data.SessionId,
        data.PlayerId,
        data.PropertyId,
    )

    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    err.Error(),
        }
    }

    e.Broker.Broadcast(log, "PropertyPurchased", fmt.Sprintf("Player %d purchased property %d", data.PlayerId, data.PropertyId))

    log.Trace().Msgf("player %d successfully purchased property %d", data.PlayerId, data.PropertyId)
    return internal.UserActionStatus{
        Status: http.StatusOK,
        Msg:    fmt.Sprintf("property purchased successfully (Ownership ID: %d)", ownershipId),
    }
}
