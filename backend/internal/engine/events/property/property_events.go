package property

import (
    "context"
    "fmt"
    "monopoly-backend/internal"
    internaldb_players "monopoly-backend/internal/db/player"
    internaldb_tiles "monopoly-backend/internal/db/tile"
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

    if currentPlayer.Id != data.PlayerId {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "it is not this player's turn",
        }
    }

    property, err := internaldb_tiles.GetPropertyData(
        log,
        ctx,
        tx,
        data.SessionId,
        data.PropertyId,
    )
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    "failed to get property data from db",
        }
    }

    if currentPlayer.Money < property.PurchaseCost {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "you can't afford this property",
        }
    }

    err = internaldb_players.UpdatePlayerMoney(log, ctx, tx, data.PlayerId, data.SessionId, currentPlayer.Money-property.PurchaseCost)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    "failed to purchase property",
        }
    }

    // try to buy property, also does ownership validation
    ownershipId, err := internaldb_tiles.CreatePropertyOwnerDB(
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

func MortgageProperty(
    ctx context.Context,
    log zerolog.Logger,
    e *internal.MonopolyEngine,
    action *internal.UserActionEvent,
    tx *pgxpool.Tx,
) internal.UserActionStatus {
    log.Trace().Msg("player attempting to mortgage property")

    data, ok := action.Data.(internal.PropertyActionData)
    if !ok {
        log.Error().Msg("invalid data format received for MortgageProperty")
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

    if currentPlayer.Id != data.PlayerId {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "it is not this player's turn",
        }
    }

    propertyGroup, propertyData, err := getMortgagePropertyData(ctx, log, tx, data)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    err.Error(),
        }
    }

    if propertyData.IsMortgaged {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "property is already mortgaged",
        }
    }

    if propertyData.PropertyType != "RAILROAD" && propertyData.PropertyType != "UTILITY" {
        for _, groupProperty := range propertyGroup {
            if groupProperty.Houses > 0 || groupProperty.HasHotel {
                return internal.UserActionStatus{
                    Status: http.StatusBadRequest,
                    Msg:    "all houses and hotels in this property set must be sold before mortgaging",
                }
            }
        }
    }

    property, err := internaldb_tiles.GetPropertyData(log, ctx, tx, data.SessionId, data.PropertyId)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    "failed to get property data from db",
        }
    }

    err = internaldb_tiles.UpdatePropertyMortgageStatus(log, ctx, tx, data.SessionId, data.PropertyId, true)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    err = internaldb_players.UpdatePlayerMoney(log, ctx, tx, data.PlayerId, data.SessionId, currentPlayer.Money+property.MortgageCost)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    propertyMortgageUpdate := internal.PropertyMortgageUpdate{
        PlayerId:    data.PlayerId,
        SessionId:   data.SessionId,
        PropertyId:  data.PropertyId,
        IsMortgaged: true,
        PlayerMoney: currentPlayer.Money + property.MortgageCost,
    }

    e.Broker.Broadcast(log, "PropertyMortgaged", propertyMortgageUpdate)

    return internal.UserActionStatus{
        Status: http.StatusOK,
        Data:   propertyMortgageUpdate,
    }
}

func UnmortgageProperty(
    ctx context.Context,
    log zerolog.Logger,
    e *internal.MonopolyEngine,
    action *internal.UserActionEvent,
    tx *pgxpool.Tx,
) internal.UserActionStatus {
    log.Trace().Msg("player attempting to unmortgage property")

    data, ok := action.Data.(internal.PropertyActionData)
    if !ok {
        log.Error().Msg("invalid data format received for UnmortgageProperty")
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

    if currentPlayer.Id != data.PlayerId {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "it is not this player's turn",
        }
    }

    _, propertyData, err := getMortgagePropertyData(ctx, log, tx, data)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    err.Error(),
        }
    }

    if !propertyData.IsMortgaged {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "property is not mortgaged",
        }
    }

    property, err := internaldb_tiles.GetPropertyData(log, ctx, tx, data.SessionId, data.PropertyId)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    "failed to get property data from db",
        }
    }

    if currentPlayer.Money < property.UnmortgageCost {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "player does not have enough money",
        }
    }

    err = internaldb_tiles.UpdatePropertyMortgageStatus(log, ctx, tx, data.SessionId, data.PropertyId, false)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    err = internaldb_players.UpdatePlayerMoney(log, ctx, tx, data.PlayerId, data.SessionId, currentPlayer.Money-property.UnmortgageCost)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    propertyMortgageUpdate := internal.PropertyMortgageUpdate{
        PlayerId:    data.PlayerId,
        SessionId:   data.SessionId,
        PropertyId:  data.PropertyId,
        IsMortgaged: false,
        PlayerMoney: currentPlayer.Money - property.UnmortgageCost,
    }

    e.Broker.Broadcast(log, "PropertyUnmortgaged", propertyMortgageUpdate)

    return internal.UserActionStatus{
        Status: http.StatusOK,
        Data:   propertyMortgageUpdate,
    }
}

func getMortgagePropertyData(
    ctx context.Context,
    log zerolog.Logger,
    tx *pgxpool.Tx,
    data internal.PropertyActionData,
) ([]internaldb_tiles.PropertyGroupData, internaldb_tiles.PropertyGroupData, error) {
    propertyGroup, err := internaldb_tiles.GetPropertyGroupData(log, ctx, tx, data.SessionId, data.PropertyId)
    if err != nil {
        return nil, internaldb_tiles.PropertyGroupData{}, err
    }

    if len(propertyGroup) == 0 {
        return nil, internaldb_tiles.PropertyGroupData{}, fmt.Errorf("property does not exist")
    }

    var propertyData internaldb_tiles.PropertyGroupData
    foundProperty := false
    for _, groupProperty := range propertyGroup {
        if groupProperty.PropertyId == data.PropertyId {
            propertyData = groupProperty
            foundProperty = true
            break
        }
    }

    if !foundProperty {
        return nil, internaldb_tiles.PropertyGroupData{}, fmt.Errorf("property does not exist")
    }

    if !propertyData.Owned || propertyData.OwnerId != data.PlayerId {
        return nil, internaldb_tiles.PropertyGroupData{}, fmt.Errorf("player does not own this property")
    }

    return propertyGroup, propertyData, nil
}
