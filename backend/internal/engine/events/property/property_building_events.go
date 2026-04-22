package property

import (
    "context"
    "fmt"
    "monopoly-backend/internal"
    internaldb_players "monopoly-backend/internal/db/player"
    internaldb_tiles "monopoly-backend/internal/db/tiles"
    turn_events "monopoly-backend/internal/engine/events/turn"
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

    currentPlayer, _, _, err := turn_events.GetCurrentPlayer(ctx, log, tx, data.SessionId)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    if currentPlayer.Id != data.PlayerId {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "it is not this player's turn",
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

    availableHouses, err := internaldb_tiles.GetAvailableHouses(log, ctx, tx, data.SessionId)
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

    err = internaldb_tiles.UpdatePropertyBuildings(log, ctx, tx, data.SessionId, data.PropertyId, propertyData.Houses+1, false)
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

    availableHouses, err = internaldb_tiles.GetAvailableHouses(log, ctx, tx, data.SessionId)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    availableHotels, err := internaldb_tiles.GetAvailableHotels(log, ctx, tx, data.SessionId)
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

    currentPlayer, _, _, err := turn_events.GetCurrentPlayer(ctx, log, tx, data.SessionId)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    if currentPlayer.Id != data.PlayerId {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "it is not this player's turn",
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

    availableHotels, err := internaldb_tiles.GetAvailableHotels(log, ctx, tx, data.SessionId)
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

    err = internaldb_tiles.UpdatePropertyBuildings(log, ctx, tx, data.SessionId, data.PropertyId, 0, true)
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

    availableHouses, err := internaldb_tiles.GetAvailableHouses(log, ctx, tx, data.SessionId)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    availableHotels, err = internaldb_tiles.GetAvailableHotels(log, ctx, tx, data.SessionId)
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

func SellHouse(
    ctx context.Context,
    log zerolog.Logger,
    e *internal.MonopolyEngine,
    action *internal.UserActionEvent,
    tx *pgxpool.Tx,
) internal.UserActionStatus {
    log.Trace().Msg("player attempting to sell house")

    data, ok := action.Data.(internal.PropertyActionData)
    if !ok {
        log.Error().Msg("invalid data format received for SellHouse")
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

    if currentPlayer.Id != data.PlayerId {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "it is not this player's turn",
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
            Msg:    "property has a hotel",
        }
    }

    if propertyData.Houses == 0 {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "property does not have a house to sell",
        }
    }

    maxLevel := getMaxPropertyLevel(propertyGroup)
    if propertyData.Houses != maxLevel {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "houses must be sold evenly",
        }
    }

    player, err := internaldb_players.GetPlayer(log, ctx, tx, data.PlayerId, data.SessionId)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    err = internaldb_tiles.UpdatePropertyBuildings(log, ctx, tx, data.SessionId, data.PropertyId, propertyData.Houses-1, false)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    err = internaldb_players.UpdatePlayerMoney(log, ctx, tx, data.PlayerId, data.SessionId, player.Money+(propertyData.HouseCost/2))
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    availableHouses, err := internaldb_tiles.GetAvailableHouses(log, ctx, tx, data.SessionId)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    availableHotels, err := internaldb_tiles.GetAvailableHotels(log, ctx, tx, data.SessionId)
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
        Houses:          propertyData.Houses - 1,
        HasHotel:        false,
        PlayerMoney:     player.Money + (propertyData.HouseCost / 2),
        AvailableHouses: availableHouses,
        AvailableHotels: availableHotels,
    }

    e.Broker.Broadcast(log, "HouseSold", propertyBuildingUpdate)

    return internal.UserActionStatus{
        Status: http.StatusOK,
        Data:   propertyBuildingUpdate,
    }
}

func SellHotel(
    ctx context.Context,
    log zerolog.Logger,
    e *internal.MonopolyEngine,
    action *internal.UserActionEvent,
    tx *pgxpool.Tx,
) internal.UserActionStatus {
    log.Trace().Msg("player attempting to sell hotel")

    data, ok := action.Data.(internal.PropertyActionData)
    if !ok {
        log.Error().Msg("invalid data format received for SellHotel")
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

    if currentPlayer.Id != data.PlayerId {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "it is not this player's turn",
        }
    }

    propertyGroup, propertyData, err := getValidatedPropertyGroup(ctx, log, tx, data)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    err.Error(),
        }
    }

    if !propertyData.HasHotel {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "property does not have a hotel to sell",
        }
    }

    maxLevel := getMaxPropertyLevel(propertyGroup)
    if getPropertyLevel(propertyData) != maxLevel {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "hotels must be sold evenly",
        }
    }

    availableHouses, err := internaldb_tiles.GetAvailableHouses(log, ctx, tx, data.SessionId)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    if availableHouses < 4 {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "there are not enough houses available in the bank",
        }
    }

    player, err := internaldb_players.GetPlayer(log, ctx, tx, data.PlayerId, data.SessionId)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    err = internaldb_tiles.UpdatePropertyBuildings(log, ctx, tx, data.SessionId, data.PropertyId, 4, false)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    err = internaldb_players.UpdatePlayerMoney(log, ctx, tx, data.PlayerId, data.SessionId, player.Money+(propertyData.HotelCost/2))
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    availableHouses, err = internaldb_tiles.GetAvailableHouses(log, ctx, tx, data.SessionId)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    availableHotels, err := internaldb_tiles.GetAvailableHotels(log, ctx, tx, data.SessionId)
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
        Houses:          4,
        HasHotel:        false,
        PlayerMoney:     player.Money + (propertyData.HotelCost / 2),
        AvailableHouses: availableHouses,
        AvailableHotels: availableHotels,
    }

    e.Broker.Broadcast(log, "HotelSold", propertyBuildingUpdate)

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
) ([]internaldb_tiles.PropertyGroupData, internaldb_tiles.PropertyGroupData, error) {
    propertyGroup, err := internaldb_tiles.GetPropertyGroupData(log, ctx, tx, data.SessionId, data.PropertyId)
    if err != nil {
        return nil, internaldb_tiles.PropertyGroupData{}, err
    }

    if len(propertyGroup) == 0 {
        return nil, internaldb_tiles.PropertyGroupData{}, fmt.Errorf("property does not exist")
    }

    propertyType := propertyGroup[0].PropertyType
    if propertyType == "RAILROAD" || propertyType == "UTILITY" {
        return nil, internaldb_tiles.PropertyGroupData{}, fmt.Errorf("property cannot have houses or hotels")
    }

    var propertyData internaldb_tiles.PropertyGroupData
    foundProperty := false
    for _, groupProperty := range propertyGroup {
        if groupProperty.PropertyId == data.PropertyId {
            propertyData = groupProperty
            foundProperty = true
        }

        if !groupProperty.Owned || groupProperty.OwnerId != data.PlayerId {
            return nil, internaldb_tiles.PropertyGroupData{}, fmt.Errorf("player does not own the full property set")
        }

        if groupProperty.IsMortgaged {
            return nil, internaldb_tiles.PropertyGroupData{}, fmt.Errorf("cannot build on a mortgaged property set")
        }
    }

    if !foundProperty {
        return nil, internaldb_tiles.PropertyGroupData{}, fmt.Errorf("property does not exist")
    }

    return propertyGroup, propertyData, nil
}

func getMinPropertyLevel(propertyGroup []internaldb_tiles.PropertyGroupData) int {
    minLevel := getPropertyLevel(propertyGroup[0])
    for _, groupProperty := range propertyGroup[1:] {
        propertyLevel := getPropertyLevel(groupProperty)
        if propertyLevel < minLevel {
            minLevel = propertyLevel
        }
    }

    return minLevel
}

func getMaxPropertyLevel(propertyGroup []internaldb_tiles.PropertyGroupData) int {
    maxLevel := getPropertyLevel(propertyGroup[0])
    for _, groupProperty := range propertyGroup[1:] {
        propertyLevel := getPropertyLevel(groupProperty)
        if propertyLevel > maxLevel {
            maxLevel = propertyLevel
        }
    }

    return maxLevel
}

func getPropertyLevel(propertyData internaldb_tiles.PropertyGroupData) int {
    if propertyData.HasHotel {
        return 5
    }

    return propertyData.Houses
}
