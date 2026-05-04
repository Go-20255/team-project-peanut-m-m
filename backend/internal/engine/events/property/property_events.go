package property

import (
	"context"
	"fmt"
	"monopoly-backend/internal"
	internaldb_players "monopoly-backend/internal/db/player"
	internaldb_tiles "monopoly-backend/internal/db/tile"
	"monopoly-backend/internal/engine/events"
	turn_events "monopoly-backend/internal/engine/events/turn"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

func getActionPlayers(
    ctx context.Context,
    log zerolog.Logger,
    tx *pgxpool.Tx,
    sessionId string,
    playerId int,
) (*internal.Player, *internal.Player, []internal.Player, int, *internal.UserActionStatus) {
    currentPlayer, actingPlayer, players, turnNumber, err := turn_events.GetActionPlayers(ctx, log, tx, sessionId, playerId)
    if err != nil {
        return nil, nil, nil, 0, &internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    if actingPlayer.Bankrupt {
        return nil, nil, nil, 0, &internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "player is bankrupt",
        }
    }

    if currentPlayer == nil {
        return nil, nil, nil, 0, &internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "there are no players in this game session",
        }
    }

    if currentPlayer.Id != playerId {
        return nil, nil, nil, 0, &internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "it is not this player's turn",
        }
    }

    return currentPlayer, actingPlayer, players, turnNumber, nil
}

// PurchaseProperty lets the current player buy the property on their current
// tile.
// It validates that it is the player's turn, confirms the tile has an
// available property, checks that the property is unowned and affordable,
// updates the player's money and ownership records, and broadcasts the result.
func PurchaseProperty(
    ctx context.Context,
    log zerolog.Logger,
    e *internal.MonopolyEngine,
    action *internal.UserActionEvent,
    tx *pgxpool.Tx,
) internal.UserActionStatus {
    log.Trace().Msg("player attempting to purchase property")

    data, ok := action.Data.(internal.SimpleActionData)
    if !ok {
        log.Error().Msg("invalid data format received for PurchaseProperty")
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "internal data format error",
        }
    }

    currentPlayer, _, _, _, status := getActionPlayers(ctx, log, tx, data.SessionId, data.PlayerId)
    if status != nil {
        return *status
    }

    if e.PendingPropertyPurchase == nil || e.PendingPropertyPurchase.PlayerId != data.PlayerId || e.PendingPropertyPurchase.SessionId != data.SessionId {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "there is no pending property purchase",
        }
    }

    tile, err := internaldb_tiles.GetTileByPosition(
        log,
        ctx,
        tx,
        data.SessionId,
        currentPlayer.Position,
        )
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    "failed to get tile data from db",
        }
    }

    property := tile.PropertyData
    if property == nil {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "current tile is not a purchasable property",
        }
    }

    // check ownership of property
    _, is_owned, err := internaldb_tiles.VerifyPropertyOwnerDB(
        log,
        ctx,
        tx,
        data.SessionId,
        property.Id,
        )
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    "failed to determine owner of this property",
        }
    }

    if is_owned {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg: "property is already owned",
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
        property.Id,
    )

    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    err.Error(),
        }
    }

    var event struct {
        PlayerId            int     `json:"player_id"`
        PropertyId          int     `json:"property_id"`
        OwnershipId         int     `json:"ownership_id"`
    }
    event.PlayerId = data.PlayerId
    event.PropertyId = property.Id
    event.OwnershipId = ownershipId

    e.PendingPropertyPurchase = nil
    e.Broker.Broadcast(log, "PropertyPurchasedEvent", event)
    events.EmitGameBoardUpdate(log, ctx, e, tx)

    log.Trace().Msgf("player %d successfully purchased property %d", data.PlayerId, property.Id)
    return internal.UserActionStatus{
        Status: http.StatusOK,
        Msg:    fmt.Sprintf("property purchased successfully (Ownership ID: %d)", ownershipId),
    }
}

func IgnorePropertyPurchase(
    ctx context.Context,
    log zerolog.Logger,
    e *internal.MonopolyEngine,
    action *internal.UserActionEvent,
    tx *pgxpool.Tx,
) internal.UserActionStatus {
    data, ok := action.Data.(internal.SimpleActionData)
    if !ok {
        log.Error().Msg("invalid data format received for IgnorePropertyPurchase")
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "internal data format error",
        }
    }

    _, _, _, _, status := getActionPlayers(ctx, log, tx, data.SessionId, data.PlayerId)
    if status != nil {
        return *status
    }

    if e.PendingPropertyPurchase == nil || e.PendingPropertyPurchase.PlayerId != data.PlayerId || e.PendingPropertyPurchase.SessionId != data.SessionId {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "there is no pending property purchase",
        }
    }

    ignoredEvent := *e.PendingPropertyPurchase
    e.PendingPropertyPurchase = nil
    e.Broker.Broadcast(log, "PropertyPurchaseIgnoredEvent", ignoredEvent)

    return internal.UserActionStatus{
        Status: http.StatusOK,
        Msg:    "property purchase ignored",
    }
}

// MortgageProperty mortgages a property owned by the current player.
// It validates turn ownership, confirms the player owns the property, checks
// that the property is not already mortgaged, and for standard property groups
// ensures all houses and hotels in the set have been sold first. On success,
// it updates the mortgage state, adds the mortgage value to the player's money,
// and broadcasts the update.
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

    currentPlayer, _, _, _, status := getActionPlayers(ctx, log, tx, data.SessionId, data.PlayerId)
    if status != nil {
        return *status
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

    e.Broker.Broadcast(log, "PropertyMortgagedEvent", propertyMortgageUpdate)

    return internal.UserActionStatus{
        Status: http.StatusOK,
        Data:   propertyMortgageUpdate,
    }
}

// UnmortgageProperty removes the mortgage from a property owned by the current
// player.
// It validates turn ownership, confirms the property is currently mortgaged,
// checks that the player can afford the unmortgage cost, updates the mortgage
// state, subtracts the cost from the player's money, and broadcasts the update.
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

    currentPlayer, _, _, _, status := getActionPlayers(ctx, log, tx, data.SessionId, data.PlayerId)
    if status != nil {
        return *status
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

    e.Broker.Broadcast(log, "PropertyUnmortgagedEvent", propertyMortgageUpdate)

    return internal.UserActionStatus{
        Status: http.StatusOK,
        Data:   propertyMortgageUpdate,
    }
}

// getMortgagePropertyData loads property-group data for a property action and
// validates that the requested property exists and is owned by the player.
// It returns the full property group along with the specific property's data so
// callers can apply mortgage rules that depend on the whole set.
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
