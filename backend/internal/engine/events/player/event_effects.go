package player
import (
	"context"
	"monopoly-backend/internal"
	"net/http"
	"github.com/jackc/pgx/v5/pgxpool"
	internaldb_players "monopoly-backend/internal/db/player"
	"github.com/rs/zerolog"
)

type CardEffectFunc func(
	ctx context.Context,
	log zerolog.Logger,
	e *internal.MonopolyEngine,
	tx *pgxpool.Tx,
	sessionId string,
	playerId int,
) internal.UserActionStatus

var CardEffects = map[int]CardEffectFunc{
	1:  advanceToGoEffect,
	2:  advanceToIllinoisAvenueEffect,
	3:  advanceToStCharlesPlaceEffect,
	4:  advanceToNearestUtilityEffect,
	5:  advanceToNearestRailroadEffect,
	6:  advanceToNearestRailroadEffect,
	7:  advanceToBoardwalkEffect,
	8:  dividendEffect,
	9:  getOutOfJailFreeEffect,
	10: goBackThreeSpacesEffect,
	11: goToJailEffect,
	12: generalRepairsEffect,
	13: speedingFineEffect,
	14: readingRailroadEffect,
	15: chairmanOfTheBoardEffect,
	16: buildingLoanMaturesEffect,

	17: advanceToGoEffect,
	18: bankErrorEffect,
	19: doctorFeeEffect,
	20: stockSaleEffect,
	21: getOutOfJailFreeEffect,
	22: goToJailEffect,
	23: holidayFundEffect,
	24: taxRefundEffect,
	25: birthdayEffect,
	26: lifeInsuranceEffect,
	27: hospitalFeesEffect,
	28: schoolFeesEffect,
	29: consultancyFeeEffect,
	30: streetRepairsEffect,
	31: beautyContestEffect,
	32: inheritanceEffect,
}

func advanceToGoEffect(ctx context.Context, log zerolog.Logger, e *internal.MonopolyEngine, tx *pgxpool.Tx, sessionId string, playerId int) internal.UserActionStatus {
	player, err := internaldb_players.GetPlayer(log, ctx, tx, playerId, sessionId)
	if err != nil {
		return internal.UserActionStatus{
			Status: http.StatusInternalServerError,
			Msg:    err.Error(),
		}
	}

	err = internaldb_players.UpdatePlayerPosition(log, ctx, tx, playerId, sessionId, 0)
	if err != nil {
		return internal.UserActionStatus{
			Status: http.StatusInternalServerError,
			Msg:    err.Error(),
		}
	}

	playerMovement := internal.PlayerMovement{
		PlayerId:    playerId,
		SessionId:   sessionId,
		OldPosition: player.Position,
		NewPosition: 0,
		PassedGo:    true,
	}
	e.Broker.Broadcast(log, "MovePlayerEvent", playerMovement)

	return SetPendingBankPayout(log, e, playerId, sessionId, 200, "Advance to Go")
}

func advanceToIllinoisAvenueEffect(ctx context.Context, log zerolog.Logger, e *internal.MonopolyEngine, tx *pgxpool.Tx, sessionId string, playerId int) internal.UserActionStatus {
	player, err := internaldb_players.GetPlayer(log, ctx, tx, playerId, sessionId)
	if err != nil {
		return internal.UserActionStatus{
			Status: http.StatusInternalServerError,
			Msg:    err.Error(),
		}
	}

	err = internaldb_players.UpdatePlayerPosition(log, ctx, tx, playerId, sessionId, 24)
	if err != nil {
		return internal.UserActionStatus{
			Status: http.StatusInternalServerError,
			Msg:    err.Error(),
		}
	}

	playerMovement := internal.PlayerMovement{
		PlayerId:    playerId,
		SessionId:   sessionId,
		OldPosition: player.Position,
		NewPosition: 24,
		PassedGo:    player.Position > 24,
	}
	e.Broker.Broadcast(log, "MovePlayerEvent", playerMovement)

	if player.Position > 24 {
		return SetPendingBankPayout(log, e, playerId, sessionId, 200, "Advance to Illinois Avenue")
	}

	return internal.UserActionStatus{Status: http.StatusOK}
}

func advanceToStCharlesPlaceEffect(ctx context.Context, log zerolog.Logger, e *internal.MonopolyEngine, tx *pgxpool.Tx, sessionId string, playerId int) internal.UserActionStatus {
	player, err := internaldb_players.GetPlayer(log, ctx, tx, playerId, sessionId)
	if err != nil {
		return internal.UserActionStatus{
			Status: http.StatusInternalServerError,
			Msg:    err.Error(),
		}
	}

	err = internaldb_players.UpdatePlayerPosition(log, ctx, tx, playerId, sessionId, 11)
	if err != nil {
		return internal.UserActionStatus{
			Status: http.StatusInternalServerError,
			Msg:    err.Error(),
		}
	}

	playerMovement := internal.PlayerMovement{
		PlayerId:    playerId,
		SessionId:   sessionId,
		OldPosition: player.Position,
		NewPosition: 11,
		PassedGo:    player.Position > 11,
	}
	e.Broker.Broadcast(log, "MovePlayerEvent", playerMovement)

	if player.Position > 11 {
		return SetPendingBankPayout(log, e, playerId, sessionId, 200, "Advance to St. Charles Place")
	}

	return internal.UserActionStatus{Status: http.StatusOK}
}

func advanceToNearestUtilityEffect(ctx context.Context, log zerolog.Logger, e *internal.MonopolyEngine, tx *pgxpool.Tx, sessionId string, playerId int) internal.UserActionStatus {
	player, err := internaldb_players.GetPlayer(log, ctx, tx, playerId, sessionId)
	if err != nil {
		return internal.UserActionStatus{
			Status: http.StatusInternalServerError,
			Msg:    err.Error(),
		}
	}

	var nearest int
	if player.Position < 12 {
		nearest = 12
	} else if player.Position < 28 {
		nearest = 28
	} else {
		nearest = 12
	}

	err = internaldb_players.UpdatePlayerPosition(log, ctx, tx, playerId, sessionId, nearest)
	if err != nil {
		return internal.UserActionStatus{
			Status: http.StatusInternalServerError,
			Msg:    err.Error(),
		}
	}

	err = internaldb_players.UpdatePlayerPosition(log, ctx, tx, playerId, sessionId, nearest)
	if err != nil {
		return internal.UserActionStatus{
			Status: http.StatusInternalServerError,
			Msg:    err.Error(),
		}
	}

	playerMovement := internal.PlayerMovement{
		PlayerId:    playerId,
		SessionId:   sessionId,
		OldPosition: player.Position,
		NewPosition: nearest,
		PassedGo:    player.Position > nearest,
	}
	e.Broker.Broadcast(log, "MovePlayerEvent", playerMovement)

	if player.Position > nearest {
		return SetPendingBankPayout(log, e, playerId, sessionId, 200, "Advance to Nearest Utility")
	}

	return internal.UserActionStatus{Status: http.StatusOK}
}

func advanceToNearestRailroadEffect(ctx context.Context, log zerolog.Logger, e *internal.MonopolyEngine, tx *pgxpool.Tx, sessionId string, playerId int) internal.UserActionStatus {
	player, err := internaldb_players.GetPlayer(log, ctx, tx, playerId, sessionId)
	if err != nil {
		return internal.UserActionStatus{
			Status: http.StatusInternalServerError,
			Msg:    err.Error(),
		}
	}

	var nearest int
	if player.Position < 5 {
		nearest = 5
	} else if player.Position < 15 {
		nearest = 15
	} else if player.Position < 25 {
		nearest = 25
	} else if player.Position < 35 {
		nearest = 35
	} else {
		nearest = 5
	}

	err = internaldb_players.UpdatePlayerPosition(log, ctx, tx, playerId, sessionId, nearest)
	if err != nil {
		return internal.UserActionStatus{
			Status: http.StatusInternalServerError,
			Msg:    err.Error(),
		}
	}

	err = internaldb_players.UpdatePlayerPosition(log, ctx, tx, playerId, sessionId, nearest)
	if err != nil {
		return internal.UserActionStatus{
			Status: http.StatusInternalServerError,
			Msg:    err.Error(),
		}
	}

	playerMovement := internal.PlayerMovement{
		PlayerId:    playerId,
		SessionId:   sessionId,
		OldPosition: player.Position,
		NewPosition: nearest,
		PassedGo:    player.Position > nearest,
	}
	e.Broker.Broadcast(log, "MovePlayerEvent", playerMovement)

	if player.Position > nearest {
		return SetPendingBankPayout(log, e, playerId, sessionId, 200, "Advance to Nearest Railroad")
	}

	return internal.UserActionStatus{Status: http.StatusOK}
}

func advanceToBoardwalkEffect(ctx context.Context, log zerolog.Logger, e *internal.MonopolyEngine, tx *pgxpool.Tx, sessionId string, playerId int) internal.UserActionStatus {
	player, err := internaldb_players.GetPlayer(log, ctx, tx, playerId, sessionId)
	if err != nil {
		return internal.UserActionStatus{
			Status: http.StatusInternalServerError,
			Msg:    err.Error(),
		}
	}

	err = internaldb_players.UpdatePlayerPosition(log, ctx, tx, playerId, sessionId, 39)
	if err != nil {
		return internal.UserActionStatus{
			Status: http.StatusInternalServerError,
			Msg:    err.Error(),
		}
	}

	playerMovement := internal.PlayerMovement{
		PlayerId:    playerId,
		SessionId:   sessionId,
		OldPosition: player.Position,
		NewPosition: 39,
		PassedGo:    player.Position > 39,
	}
	e.Broker.Broadcast(log, "MovePlayerEvent", playerMovement)

	if player.Position > 39 {
		return SetPendingBankPayout(log, e, playerId, sessionId, 200, "Advance to Boardwalk")
	}

	return internal.UserActionStatus{Status: http.StatusOK}
}

func dividendEffect(ctx context.Context, log zerolog.Logger, e *internal.MonopolyEngine, tx *pgxpool.Tx, sessionId string, playerId int) internal.UserActionStatus {
	return SetPendingBankPayout(log, e, playerId, sessionId, 50, "Bank pays you dividend")
}

func getOutOfJailFreeEffect(ctx context.Context, log zerolog.Logger, e *internal.MonopolyEngine, tx *pgxpool.Tx, sessionId string, playerId int) internal.UserActionStatus {
	// already handled in event_cards_db.go
	return internal.UserActionStatus{Status: http.StatusOK}
}

func goBackThreeSpacesEffect(ctx context.Context, log zerolog.Logger, e *internal.MonopolyEngine, tx *pgxpool.Tx, sessionId string, playerId int) internal.UserActionStatus {
	player, err := internaldb_players.GetPlayer(log, ctx, tx, playerId, sessionId)
	if err != nil {
		return internal.UserActionStatus{
			Status: http.StatusInternalServerError,
			Msg:    err.Error(),
		}
	}

	newPos := player.Position - 3
	if newPos < 0 {
		newPos += 40
	}

	err = internaldb_players.UpdatePlayerPosition(log, ctx, tx, playerId, sessionId, newPos)
	if err != nil {
		return internal.UserActionStatus{
			Status: http.StatusInternalServerError,
			Msg:    err.Error(),
		}
	}

	playerMovement := internal.PlayerMovement{
		PlayerId:    playerId,
		SessionId:   sessionId,
		OldPosition: player.Position,
		NewPosition: newPos,
		PassedGo:    false,
	}
	e.Broker.Broadcast(log, "MovePlayerEvent", playerMovement)

	return internal.UserActionStatus{Status: http.StatusOK}
}

func goToJailEffect(ctx context.Context, log zerolog.Logger, e *internal.MonopolyEngine, tx *pgxpool.Tx, sessionId string, playerId int) internal.UserActionStatus {
	err := internaldb_players.UpdatePlayerPositionAndJailed(log, ctx, tx, playerId, sessionId, 10, 1)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    return internal.UserActionStatus{Status: http.StatusOK}
}

func generalRepairsEffect(ctx context.Context, log zerolog.Logger, e *internal.MonopolyEngine, tx *pgxpool.Tx, sessionId string, playerId int) internal.UserActionStatus {
	properties, err := internaldb_players.GetPlayerOwnedProperties(log, ctx, tx, playerId, sessionId)
	if err != nil {
		return internal.UserActionStatus{Status: http.StatusInternalServerError, Msg: err.Error()}
	}

	repairCost := 0
	for _, p := range properties {
		if p.HasHotel {
			repairCost += 100
		} else {
			repairCost += p.Houses * 25
		}
	}

	if repairCost > 0 {
		return SetPendingBankPayment(log, e, playerId, sessionId, repairCost, "General Repairs")
	}

	return internal.UserActionStatus{Status: http.StatusOK}
}

func speedingFineEffect(ctx context.Context, log zerolog.Logger, e *internal.MonopolyEngine, tx *pgxpool.Tx, sessionId string, playerId int) internal.UserActionStatus {
	return SetPendingBankPayment(log, e, playerId, sessionId, 15, "Speeding Fine")
}

func readingRailroadEffect(ctx context.Context, log zerolog.Logger, e *internal.MonopolyEngine, tx *pgxpool.Tx, sessionId string, playerId int) internal.UserActionStatus {
	player, err := internaldb_players.GetPlayer(log, ctx, tx, playerId, sessionId)
	if err != nil {
		return internal.UserActionStatus{Status: http.StatusInternalServerError, Msg: err.Error()}
	}

	err = internaldb_players.UpdatePlayerPosition(log, ctx, tx, playerId, sessionId, 5)
	if err != nil {
		return internal.UserActionStatus{Status: http.StatusInternalServerError, Msg: err.Error()}
	}

	playerMovement := internal.PlayerMovement{
		PlayerId:    playerId,
		SessionId:   sessionId,
		OldPosition: player.Position,
		NewPosition: 5,
		PassedGo:    player.Position > 5,
	}
	e.Broker.Broadcast(log, "MovePlayerEvent", playerMovement)

	if player.Position > 5 {
		return SetPendingBankPayout(log, e, playerId, sessionId, 200, "Advance to Reading Railroad")
	}

	return internal.UserActionStatus{Status: http.StatusOK}
}

func chairmanOfTheBoardEffect(ctx context.Context, log zerolog.Logger, e *internal.MonopolyEngine, tx *pgxpool.Tx, sessionId string, playerId int) internal.UserActionStatus {
	players, err := internaldb_players.GetPlayersInSession(log, ctx, tx, sessionId)
	if err != nil {
		return internal.UserActionStatus{Status: http.StatusInternalServerError, Msg: err.Error()}
	}

	totalPayment := (len(players) - 1) * 50
	if totalPayment <= 0 {
		return internal.UserActionStatus{Status: http.StatusOK}
	}

	return SetPendingBankPayment(log, e, playerId, sessionId, totalPayment, "Chairman of the Board")
}

func buildingLoanMaturesEffect(ctx context.Context, log zerolog.Logger, e *internal.MonopolyEngine, tx *pgxpool.Tx, sessionId string, playerId int) internal.UserActionStatus {
	return SetPendingBankPayout(log, e, playerId, sessionId, 150, "Building Loan Matures")
}

func bankErrorEffect(ctx context.Context, log zerolog.Logger, e *internal.MonopolyEngine, tx *pgxpool.Tx, sessionId string, playerId int) internal.UserActionStatus {
	return internal.UserActionStatus{Status: http.StatusOK}
}

func doctorFeeEffect(ctx context.Context, log zerolog.Logger, e *internal.MonopolyEngine, tx *pgxpool.Tx, sessionId string, playerId int) internal.UserActionStatus {
	return internal.UserActionStatus{Status: http.StatusOK}
}

func stockSaleEffect(ctx context.Context, log zerolog.Logger, e *internal.MonopolyEngine, tx *pgxpool.Tx, sessionId string, playerId int) internal.UserActionStatus {
	return internal.UserActionStatus{Status: http.StatusOK}
}

func holidayFundEffect(ctx context.Context, log zerolog.Logger, e *internal.MonopolyEngine, tx *pgxpool.Tx, sessionId string, playerId int) internal.UserActionStatus {
	return internal.UserActionStatus{Status: http.StatusOK}
}

func taxRefundEffect(ctx context.Context, log zerolog.Logger, e *internal.MonopolyEngine, tx *pgxpool.Tx, sessionId string, playerId int) internal.UserActionStatus {
	return internal.UserActionStatus{Status: http.StatusOK}
}

func birthdayEffect(ctx context.Context, log zerolog.Logger, e *internal.MonopolyEngine, tx *pgxpool.Tx, sessionId string, playerId int) internal.UserActionStatus {
	return internal.UserActionStatus{Status: http.StatusOK}
}

func lifeInsuranceEffect(ctx context.Context, log zerolog.Logger, e *internal.MonopolyEngine, tx *pgxpool.Tx, sessionId string, playerId int) internal.UserActionStatus {
	return internal.UserActionStatus{Status: http.StatusOK}
}

func hospitalFeesEffect(ctx context.Context, log zerolog.Logger, e *internal.MonopolyEngine, tx *pgxpool.Tx, sessionId string, playerId int) internal.UserActionStatus {
	return internal.UserActionStatus{Status: http.StatusOK}
}

func schoolFeesEffect(ctx context.Context, log zerolog.Logger, e *internal.MonopolyEngine, tx *pgxpool.Tx, sessionId string, playerId int) internal.UserActionStatus {
	return internal.UserActionStatus{Status: http.StatusOK}
}

func consultancyFeeEffect(ctx context.Context, log zerolog.Logger, e *internal.MonopolyEngine, tx *pgxpool.Tx, sessionId string, playerId int) internal.UserActionStatus {
	return internal.UserActionStatus{Status: http.StatusOK}
}

func streetRepairsEffect(ctx context.Context, log zerolog.Logger, e *internal.MonopolyEngine, tx *pgxpool.Tx, sessionId string, playerId int) internal.UserActionStatus {
	return internal.UserActionStatus{Status: http.StatusOK}
}

func beautyContestEffect(ctx context.Context, log zerolog.Logger, e *internal.MonopolyEngine, tx *pgxpool.Tx, sessionId string, playerId int) internal.UserActionStatus {
	return internal.UserActionStatus{Status: http.StatusOK}
}

func inheritanceEffect(ctx context.Context, log zerolog.Logger, e *internal.MonopolyEngine, tx *pgxpool.Tx, sessionId string, playerId int) internal.UserActionStatus {
	return internal.UserActionStatus{Status: http.StatusOK}
}