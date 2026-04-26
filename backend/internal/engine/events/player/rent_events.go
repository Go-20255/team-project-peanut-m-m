package player

import (
	"context"
	"monopoly-backend/internal"
	internaldb_players "monopoly-backend/internal/db/player"
	internaldb_tiles "monopoly-backend/internal/db/tile"
	"monopoly-backend/internal/engine/events"
	turn_events "monopoly-backend/internal/engine/events/turn"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

func getFullSetRent(ctx context.Context, log zerolog.Logger, tx *pgxpool.Tx, sessionId string, propertyId int, ownerId int) (bool, error) {
    propertyGroup, err := internaldb_tiles.GetPropertyGroupData(log, ctx, tx, sessionId, propertyId)
    if err != nil {
        return false, err
    }

    for _, propertyData := range propertyGroup {
        if !propertyData.Owned || propertyData.OwnerId != ownerId {
            return false, nil
        }
    }

    return true, nil
}

func getRentAmount(
    ctx context.Context,
    log zerolog.Logger,
    tx *pgxpool.Tx,
    sessionId string,
    tileData internaldb_tiles.RentTileData,
    diceTotal int,
) (int, error) {
    if tileData.PropertyType == "RAILROAD" {
        railroadCount, err := internaldb_tiles.GetOwnedPropertyTypeCount(log, ctx, tx, sessionId, tileData.OwnerId, "RAILROAD")
        if err != nil {
            return 0, err
        }

        if railroadCount == 1 {
            return 25, nil
        }
        if railroadCount == 2 {
            return 50, nil
        }
        if railroadCount == 3 {
            return 100, nil
        }
        if railroadCount >= 4 {
            return 200, nil
        }

        return 0, nil
    }

    if tileData.PropertyType == "UTILITY" {
        utilityCount, err := internaldb_tiles.GetOwnedPropertyTypeCount(log, ctx, tx, sessionId, tileData.OwnerId, "UTILITY")
        if err != nil {
            return 0, err
        }

        if utilityCount >= 2 {
            return diceTotal * 10, nil
        }

        if utilityCount == 1 {
            return diceTotal * 4, nil
        }

        return 0, nil
    }

    if tileData.HasHotel {
        return tileData.HotelRent, nil
    }

    if tileData.Houses == 4 {
        return tileData.FourHouseRent, nil
    }

    if tileData.Houses == 3 {
        return tileData.ThreeHouseRent, nil
    }

    if tileData.Houses == 2 {
        return tileData.TwoHouseRent, nil
    }

    if tileData.Houses == 1 {
        return tileData.OneHouseRent, nil
    }

    fullSetRent, err := getFullSetRent(ctx, log, tx, sessionId, tileData.PropertyId, tileData.OwnerId)
    if err != nil {
        return 0, err
    }

    if fullSetRent {
        return tileData.ColorSetRent, nil
    }

    return tileData.BaseRent, nil
}

func PayRent(
    ctx context.Context,
    log zerolog.Logger,
    e *internal.MonopolyEngine,
    action *internal.UserActionEvent,
    tx *pgxpool.Tx,
) internal.UserActionStatus {
    data := action.Data.(internal.RentPaymentActionData)

    if e.PendingRent == nil {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "there is no pending rent payment",
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

    if currentPlayer.Id != data.FromPlayerId {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "it is not this player's turn",
        }
    }

    if e.PendingRent.FromPlayerId != data.FromPlayerId {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "rent payer is incorrect",
        }
    }

    if e.PendingRent.ToPlayerId != data.ToPlayerId {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "rent recipient is incorrect",
        }
    }

    if e.PendingRent.Amount != data.Amount {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "rent amount is incorrect",
        }
    }

    payer, err := internaldb_players.GetPlayer(log, ctx, tx, data.FromPlayerId, data.SessionId)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    if payer.Position != e.PendingRent.Position {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "player is not on the correct tile to pay rent",
        }
    }

    tileData, err := internaldb_tiles.GetRentTileData(log, ctx, tx, data.SessionId, payer.Position)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    if !tileData.HasProperty || !tileData.Owned || tileData.OwnerId != data.ToPlayerId {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "rent recipient is incorrect",
        }
    }

    if tileData.IsMortgaged {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "rent cannot be paid on a mortgaged property",
        }
    }

    if tileData.OwnerId == data.FromPlayerId {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "player does not owe rent on this property",
        }
    }

    expectedRentAmount, err := getRentAmount(ctx, log, tx, data.SessionId, tileData, e.PendingRent.DiceTotal)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    if expectedRentAmount != data.Amount {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "rent amount is incorrect",
        }
    }

    recipient, err := internaldb_players.GetPlayer(log, ctx, tx, data.ToPlayerId, data.SessionId)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    if payer.Money < data.Amount {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "player does not have enough money",
        }
    }

    err = internaldb_players.UpdatePlayerMoney(log, ctx, tx, data.FromPlayerId, data.SessionId, payer.Money-data.Amount)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    err = internaldb_players.UpdatePlayerMoney(log, ctx, tx, data.ToPlayerId, data.SessionId, recipient.Money+data.Amount)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    rentPayment := internal.RentPayment{
        FromPlayerId:    data.FromPlayerId,
        ToPlayerId:      data.ToPlayerId,
        SessionId:       data.SessionId,
        PropertyId:      tileData.PropertyId,
        Amount:          data.Amount,
        FromPlayerMoney: payer.Money - data.Amount,
        ToPlayerMoney:   recipient.Money + data.Amount,
    }

    e.PendingRent = nil
    e.Broker.Broadcast(log, "RentPaidEvent", rentPayment)
    events.EmitGameBoardUpdate(log, ctx, e, tx)

    return internal.UserActionStatus{
        Status: http.StatusOK,
        Data:   rentPayment,
    }
}
