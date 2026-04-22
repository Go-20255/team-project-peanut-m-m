package property

import (
	"context"
	"fmt"
	"monopoly-backend/internal"
	internaldb_players "monopoly-backend/internal/db/player"
	internaldb_properties "monopoly-backend/internal/db/property"
	turn_events "monopoly-backend/internal/engine/events/turn"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

func PurchaseProperty(
    ctx context.Context,
    log zerolog.Logger,
    e *internal.MonopolyEngine,
    action *internal.UserActionEvent,
    tx *pgxpool.Tx,
) internal.UserActionStatus {
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

    currentPlayer, _, _, err := turn_events.GetCurrentPlayer(ctx, log, tx, data.SessionId)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    if currentPlayer == nil {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "there are no players in this game session",
        }
    }

    // TODO undo this temporary comment out :) - Michael

    //if currentPlayer.Id != data.PlayerId {
        //return internal.UserActionStatus{
            //Status: http.StatusBadRequest,
            //Msg:    "it is not this player's turn",
        //}
    //}

    property, err := internaldb_properties.GetPropertyData(
        log,
        ctx,
        tx,
        data.SessionId,
        data.PropertyId,
        )
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg: "failed to get property data from db",
        }
    }

    if currentPlayer.Money < property.PurchaseCost {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg: "you can't afford this property",
        }
    }

    err = internaldb_players.UpdatePlayerMoney(log, ctx, tx, data.PlayerId, data.SessionId, currentPlayer.Money - property.PurchaseCost)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg: "failed to purchase property",
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
