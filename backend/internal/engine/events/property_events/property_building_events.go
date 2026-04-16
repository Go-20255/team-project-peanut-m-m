package property_events

import (
    "context"
    "fmt"
    "monopoly-backend/internal"
    internaldb_players "monopoly-backend/internal/db/players"
    internaldb_properties "monopoly-backend/internal/db/properties"
    "net/http"

    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/rs/zerolog"
)

func PurchaseHouse(
    ctx context.Context,
    log zerolog.Logger,
    e *internal.MonopolyEngine,
    action *internal.UserActionEvent,
    tx *pgxpool.Tx,
) internal.UserActionStatus {
    log.Trace().Msg("player attempting to purchase house")

    data, ok := action.Data.(internal.PropertyActionData)
    if !ok {
        log.Error().Msg("invalid data format received for PurchaseHouse")
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "internal data format error",
        }
    }

    propertyGroup, propertyData, err := getValidatedPropertyGroup(ctx, log, tx, data)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    err.Error(),
        }
    }

    if propertyData.HasHotel {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "property already has a hotel",
        }
    }

    if propertyData.Houses >= 4 {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "property already has four houses",
        }
    }

    minLevel := getMinPropertyLevel(propertyGroup)
    if propertyData.Houses != minLevel {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "houses must be built evenly",
        }
    }

    availableHouses, err := internaldb_properties.GetAvailableHouses(log, ctx, tx, data.SessionId)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    if availableHouses < 1 {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "there are no houses available in the bank",
        }
    }

    player, err := internaldb_players.GetPlayer(log, ctx, tx, data.PlayerId, data.SessionId)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    if player.Money < propertyData.HouseCost {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "player does not have enough money",
        }
    }

    err = internaldb_properties.UpdatePropertyBuildings(log, ctx, tx, data.SessionId, data.PropertyId, propertyData.Houses+1, false)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    err = internaldb_players.UpdatePlayerMoney(log, ctx, tx, data.PlayerId, data.SessionId, player.Money-propertyData.HouseCost)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    availableHouses, err = internaldb_properties.GetAvailableHouses(log, ctx, tx, data.SessionId)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    availableHotels, err := internaldb_properties.GetAvailableHotels(log, ctx, tx, data.SessionId)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    propertyBuildingUpdate := internal.PropertyBuildingUpdate{
        PlayerId:        data.PlayerId,
        SessionId:       data.SessionId,
        PropertyId:      data.PropertyId,
        Houses:          propertyData.Houses + 1,
        HasHotel:        false,
        PlayerMoney:     player.Money - propertyData.HouseCost,
        AvailableHouses: availableHouses,
        AvailableHotels: availableHotels,
    }

    e.Broker.Broadcast(log, "HousePurchased", propertyBuildingUpdate)

    return internal.UserActionStatus{
        Status: http.StatusOK,
        Data:   propertyBuildingUpdate,
    }
}

func PurchaseHotel(
    ctx context.Context,
    log zerolog.Logger,
    e *internal.MonopolyEngine,
    action *internal.UserActionEvent,
    tx *pgxpool.Tx,
) internal.UserActionStatus {
    log.Trace().Msg("player attempting to purchase hotel")

    data, ok := action.Data.(internal.PropertyActionData)
    if !ok {
        log.Error().Msg("invalid data format received for PurchaseHotel")
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "internal data format error",
        }
    }

    propertyGroup, propertyData, err := getValidatedPropertyGroup(ctx, log, tx, data)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    err.Error(),
        }
    }

    if propertyData.HasHotel {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "property already has a hotel",
        }
    }

    if propertyData.Houses != 4 {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "property must have four houses before buying a hotel",
        }
    }

    minLevel := getMinPropertyLevel(propertyGroup)
    if propertyData.Houses != minLevel {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "hotels must be built evenly",
        }
    }

    availableHotels, err := internaldb_properties.GetAvailableHotels(log, ctx, tx, data.SessionId)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    if availableHotels < 1 {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "there are no hotels available in the bank",
        }
    }

    player, err := internaldb_players.GetPlayer(log, ctx, tx, data.PlayerId, data.SessionId)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    if player.Money < propertyData.HotelCost {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "player does not have enough money",
        }
    }

    err = internaldb_properties.UpdatePropertyBuildings(log, ctx, tx, data.SessionId, data.PropertyId, 0, true)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    err = internaldb_players.UpdatePlayerMoney(log, ctx, tx, data.PlayerId, data.SessionId, player.Money-propertyData.HotelCost)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    availableHouses, err := internaldb_properties.GetAvailableHouses(log, ctx, tx, data.SessionId)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    availableHotels, err = internaldb_properties.GetAvailableHotels(log, ctx, tx, data.SessionId)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    propertyBuildingUpdate := internal.PropertyBuildingUpdate{
        PlayerId:        data.PlayerId,
        SessionId:       data.SessionId,
        PropertyId:      data.PropertyId,
        Houses:          0,
        HasHotel:        true,
        PlayerMoney:     player.Money - propertyData.HotelCost,
        AvailableHouses: availableHouses,
        AvailableHotels: availableHotels,
    }

    e.Broker.Broadcast(log, "HotelPurchased", propertyBuildingUpdate)

    return internal.UserActionStatus{
        Status: http.StatusOK,
        Data:   propertyBuildingUpdate,
    }
}

func getValidatedPropertyGroup(
    ctx context.Context,
    log zerolog.Logger,
    tx *pgxpool.Tx,
    data internal.PropertyActionData,
) ([]internaldb_properties.PropertyGroupData, internaldb_properties.PropertyGroupData, error) {
    propertyGroup, err := internaldb_properties.GetPropertyGroupData(log, ctx, tx, data.SessionId, data.PropertyId)
    if err != nil {
        return nil, internaldb_properties.PropertyGroupData{}, err
    }

    if len(propertyGroup) == 0 {
        return nil, internaldb_properties.PropertyGroupData{}, fmt.Errorf("property does not exist")
    }

    propertyType := propertyGroup[0].PropertyType
    if propertyType == "RAILROAD" || propertyType == "UTILITY" {
        return nil, internaldb_properties.PropertyGroupData{}, fmt.Errorf("property cannot have houses or hotels")
    }

    var propertyData internaldb_properties.PropertyGroupData
    foundProperty := false
    for _, groupProperty := range propertyGroup {
        if groupProperty.PropertyId == data.PropertyId {
            propertyData = groupProperty
            foundProperty = true
        }

        if !groupProperty.Owned || groupProperty.OwnerId != data.PlayerId {
            return nil, internaldb_properties.PropertyGroupData{}, fmt.Errorf("player does not own the full property set")
        }

        if groupProperty.IsMortgaged {
            return nil, internaldb_properties.PropertyGroupData{}, fmt.Errorf("cannot build on a mortgaged property set")
        }
    }

    if !foundProperty {
        return nil, internaldb_properties.PropertyGroupData{}, fmt.Errorf("property does not exist")
    }

    return propertyGroup, propertyData, nil
}

func getMinPropertyLevel(propertyGroup []internaldb_properties.PropertyGroupData) int {
    minLevel := getPropertyLevel(propertyGroup[0])
    for _, groupProperty := range propertyGroup[1:] {
        propertyLevel := getPropertyLevel(groupProperty)
        if propertyLevel < minLevel {
            minLevel = propertyLevel
        }
    }

    return minLevel
}

func getPropertyLevel(propertyData internaldb_properties.PropertyGroupData) int {
    if propertyData.HasHotel {
        return 5
    }

    return propertyData.Houses
}
