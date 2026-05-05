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
	5:  advanceToNearestRailroadBackpackEffect,
	6:  advanceToNearestRailroadLaptopEffect,
	7:  advanceToBoardwalkEffect,
	8:  dividendEffect,
	9:  getOutOfJailFreeEffect,
	10: goToJailEffect,
	11: generalRepairsEffect,
	12: speedingFineEffect,
	13: readingRailroadEffect,
	14: chairmanOfTheBoardEffect,
	15: buildingLoanMaturesEffect,

	16: advanceToGoEffect,
	17: bankErrorEffect,
	18: doctorFeeEffect,
	19: stockSaleEffect,
	20: getOutOfJailFreeEffect,
	21: goToJailEffect,
	22: holidayFundEffect,
	23: taxRefundEffect,
	24: birthdayEffect,
	25: lifeInsuranceEffect,
	26: hospitalFeesEffect,
	27: schoolFeesEffect,
	28: consultancyFeeEffect,
	29: streetRepairsEffect,
	30: beautyContestEffect,
	31: inheritanceEffect,
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

	e.TempStore["card_movement"] = internal.PlayerMovement{
		PlayerId:    playerId,
		SessionId:   sessionId,
		OldPosition: player.Position,
		NewPosition: 0,
		PassedGo:    true,
		FromCard:    true,
	}

	return SetPendingBankPayout(log, e, playerId, sessionId, 200, "Financial aid")
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

	e.TempStore["card_movement"] = internal.PlayerMovement{
		PlayerId:    playerId,
		SessionId:   sessionId,
		OldPosition: player.Position,
		NewPosition: 24,
		PassedGo:    player.Position > 24,
		FromCard:    true,
	}

	if player.Position > 24 {
		return SetPendingBankPayout(log, e, playerId, sessionId, 200, "Work on a game")
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

	e.TempStore["card_movement"] = internal.PlayerMovement{
		PlayerId:    playerId,
		SessionId:   sessionId,
		OldPosition: player.Position,
		NewPosition: 11,
		PassedGo:    player.Position > 11,
		FromCard:    true,
	}

	if player.Position > 11 {
		return SetPendingBankPayout(log, e, playerId, sessionId, 200, "Redeem a free coffee token")
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

	e.TempStore["special_utility_rent"] = true

	e.TempStore["card_movement"] = internal.PlayerMovement{
		PlayerId:    playerId,
		SessionId:   sessionId,
		OldPosition: player.Position,
		NewPosition: nearest,
		PassedGo:    player.Position > nearest,
		FromCard:    true,
	}

	if player.Position > nearest {
		return SetPendingBankPayout(log, e, playerId, sessionId, 200, "Help FMS fix something")
	}

	return internal.UserActionStatus{Status: http.StatusOK}
}

func advanceToNearestRailroadEffect(
	ctx context.Context,
	log zerolog.Logger,
	e *internal.MonopolyEngine,
	tx *pgxpool.Tx,
	sessionId string,
	playerId int,
	reason string,
) internal.UserActionStatus {
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

	e.TempStore["special_railroad_rent"] = true

	e.TempStore["card_movement"] = internal.PlayerMovement{
		PlayerId:    playerId,
		SessionId:   sessionId,
		OldPosition: player.Position,
		NewPosition: nearest,
		PassedGo:    player.Position > nearest,
		FromCard:    true,
	}

	if player.Position > nearest {
		return SetPendingBankPayout(log, e, playerId, sessionId, 200, reason)
	}

	return internal.UserActionStatus{Status: http.StatusOK}
}

func advanceToNearestRailroadBackpackEffect(ctx context.Context, log zerolog.Logger, e *internal.MonopolyEngine, tx *pgxpool.Tx, sessionId string, playerId int) internal.UserActionStatus {
	return advanceToNearestRailroadEffect(ctx, log, e, tx, sessionId, playerId, "Left your backpack in the car")
}

func advanceToNearestRailroadLaptopEffect(ctx context.Context, log zerolog.Logger, e *internal.MonopolyEngine, tx *pgxpool.Tx, sessionId string, playerId int) internal.UserActionStatus {
	return advanceToNearestRailroadEffect(ctx, log, e, tx, sessionId, playerId, "Left your laptop in the car")
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

	e.TempStore["card_movement"] = internal.PlayerMovement{
		PlayerId:    playerId,
		SessionId:   sessionId,
		OldPosition: player.Position,
		NewPosition: 39,
		PassedGo:    player.Position > 39,
		FromCard:    true,
	}

	if player.Position > 39 {
		return SetPendingBankPayout(log, e, playerId, sessionId, 200, "Go to a performance at the Munson Music Loft")
	}

	return internal.UserActionStatus{Status: http.StatusOK}
}

func dividendEffect(ctx context.Context, log zerolog.Logger, e *internal.MonopolyEngine, tx *pgxpool.Tx, sessionId string, playerId int) internal.UserActionStatus {
	return SetPendingBankPayout(log, e, playerId, sessionId, 50, "Won a drawing for filling out a survey")
}

func getOutOfJailFreeEffect(ctx context.Context, log zerolog.Logger, e *internal.MonopolyEngine, tx *pgxpool.Tx, sessionId string, playerId int) internal.UserActionStatus {
	// already handled in event_cards_db.go
	return internal.UserActionStatus{Status: http.StatusOK}
}

func goToJailEffect(ctx context.Context, log zerolog.Logger, e *internal.MonopolyEngine, tx *pgxpool.Tx, sessionId string, playerId int) internal.UserActionStatus {
	player, err := internaldb_players.GetPlayer(log, ctx, tx, playerId, sessionId)
	if err != nil {
		return internal.UserActionStatus{
			Status: http.StatusInternalServerError,
			Msg:    err.Error(),
		}
	}

	err = internaldb_players.UpdatePlayerPositionAndJailed(log, ctx, tx, playerId, sessionId, 10, 1)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    e.TempStore["card_movement"] = internal.PlayerMovement{
        PlayerId:    playerId,
        SessionId:   sessionId,
        OldPosition: player.Position,
        NewPosition: 10,
        PassedGo:    false,
        FromCard:    true,
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
		return SetPendingBankPayment(log, e, playerId, sessionId, repairCost, "Make general repairs on all your property")
	}

	return internal.UserActionStatus{Status: http.StatusOK}
}

func speedingFineEffect(ctx context.Context, log zerolog.Logger, e *internal.MonopolyEngine, tx *pgxpool.Tx, sessionId string, playerId int) internal.UserActionStatus {
	return SetPendingBankPayment(log, e, playerId, sessionId, 15, "Buy a coffee from Midnight Oil")
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

	e.TempStore["card_movement"] = internal.PlayerMovement{
		PlayerId:    playerId,
		SessionId:   sessionId,
		OldPosition: player.Position,
		NewPosition: 5,
		PassedGo:    player.Position > 5,
		FromCard:    true,
	}

	if player.Position > 5 {
		return SetPendingBankPayout(log, e, playerId, sessionId, 200, "Print your resume from DSP")
	}

	return internal.UserActionStatus{Status: http.StatusOK}
}

func chairmanOfTheBoardEffect(ctx context.Context, log zerolog.Logger, e *internal.MonopolyEngine, tx *pgxpool.Tx, sessionId string, playerId int) internal.UserActionStatus {
	return SetPendingPlayerExchange(log, e, playerId, sessionId, 50, "Chairman of the Board", true)
}

func buildingLoanMaturesEffect(ctx context.Context, log zerolog.Logger, e *internal.MonopolyEngine, tx *pgxpool.Tx, sessionId string, playerId int) internal.UserActionStatus {
	return SetPendingBankPayout(log, e, playerId, sessionId, 150, "Your biweekly paycheck from your on-campus job just hit")
}

func bankErrorEffect(ctx context.Context, log zerolog.Logger, e *internal.MonopolyEngine, tx *pgxpool.Tx, sessionId string, playerId int) internal.UserActionStatus {
	return SetPendingBankPayout(log, e, playerId, sessionId, 200, "Scholarship money")
}

func doctorFeeEffect(ctx context.Context, log zerolog.Logger, e *internal.MonopolyEngine, tx *pgxpool.Tx, sessionId string, playerId int) internal.UserActionStatus {
	return SetPendingBankPayment(log, e, playerId, sessionId, 50, "Buy merch from the Digital Den")
}

func stockSaleEffect(ctx context.Context, log zerolog.Logger, e *internal.MonopolyEngine, tx *pgxpool.Tx, sessionId string, playerId int) internal.UserActionStatus {
	return SetPendingBankPayout(log, e, playerId, sessionId, 50, "Win a survey raffle")
}

func holidayFundEffect(ctx context.Context, log zerolog.Logger, e *internal.MonopolyEngine, tx *pgxpool.Tx, sessionId string, playerId int) internal.UserActionStatus {
	return SetPendingBankPayout(log, e, playerId, sessionId, 100, "Your biweekly paycheck from your on-campus job just hit")
}

func taxRefundEffect(ctx context.Context, log zerolog.Logger, e *internal.MonopolyEngine, tx *pgxpool.Tx, sessionId string, playerId int) internal.UserActionStatus {
	return SetPendingBankPayout(log, e, playerId, sessionId, 20, "Find some money on the ground")
}

func birthdayEffect(ctx context.Context, log zerolog.Logger, e *internal.MonopolyEngine, tx *pgxpool.Tx, sessionId string, playerId int) internal.UserActionStatus {
	return SetPendingPlayerExchange(log, e, playerId, sessionId, 10, "Birthday", false)
}

func lifeInsuranceEffect(ctx context.Context, log zerolog.Logger, e *internal.MonopolyEngine, tx *pgxpool.Tx, sessionId string, playerId int) internal.UserActionStatus {
	return SetPendingBankPayout(log, e, playerId, sessionId, 100, "Tiger Bucks reloaded")
}

func hospitalFeesEffect(ctx context.Context, log zerolog.Logger, e *internal.MonopolyEngine, tx *pgxpool.Tx, sessionId string, playerId int) internal.UserActionStatus {
	return SetPendingBankPayment(log, e, playerId, sessionId, 100, "Student activity fee")
}

func schoolFeesEffect(ctx context.Context, log zerolog.Logger, e *internal.MonopolyEngine, tx *pgxpool.Tx, sessionId string, playerId int) internal.UserActionStatus {
	return SetPendingBankPayment(log, e, playerId, sessionId, 50, "Student health fee")
}

func consultancyFeeEffect(ctx context.Context, log zerolog.Logger, e *internal.MonopolyEngine, tx *pgxpool.Tx, sessionId string, playerId int) internal.UserActionStatus {
	return SetPendingBankPayout(log, e, playerId, sessionId, 25, "Win a survey raffle")
}

func streetRepairsEffect(ctx context.Context, log zerolog.Logger, e *internal.MonopolyEngine, tx *pgxpool.Tx, sessionId string, playerId int) internal.UserActionStatus {
	properties, err := internaldb_players.GetPlayerOwnedProperties(log, ctx, tx, playerId, sessionId)
	if err != nil {
		return internal.UserActionStatus{Status: http.StatusInternalServerError, Msg: err.Error()}
	}

	repairCost := 0
	for _, p := range properties {
		if p.HasHotel {
			repairCost += 115
		} else {
			repairCost += p.Houses * 40
		}
	}

	if repairCost > 0 {
		return SetPendingBankPayment(log, e, playerId, sessionId, repairCost, "You are assessed for street repairs")
	}

	return internal.UserActionStatus{Status: http.StatusOK}
}

func beautyContestEffect(ctx context.Context, log zerolog.Logger, e *internal.MonopolyEngine, tx *pgxpool.Tx, sessionId string, playerId int) internal.UserActionStatus {
	return SetPendingBankPayout(log, e, playerId, sessionId, 10, "You win second place in Tiger Tank")
}

func inheritanceEffect(ctx context.Context, log zerolog.Logger, e *internal.MonopolyEngine, tx *pgxpool.Tx, sessionId string, playerId int) internal.UserActionStatus {
	return SetPendingBankPayout(log, e, playerId, sessionId, 100, "Your biweekly paycheck from your on-campus job just hit")
}
