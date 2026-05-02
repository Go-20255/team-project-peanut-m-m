package player

import (
    "context"
    "fmt"
    "math/rand/v2"
    "monopoly-backend/internal"
    internaldb_game_state "monopoly-backend/internal/db/game_state"
    internaldb_players "monopoly-backend/internal/db/player"
    internaldb_tiles "monopoly-backend/internal/db/tile"
	internaldb_event_cards "monopoly-backend/internal/db/event_cards"
    "monopoly-backend/internal/engine/events"
    turn_events "monopoly-backend/internal/engine/events/turn"
    "net/http"
    "sort"

    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/rs/zerolog"
)

func SetPendingBankPayment(
    log zerolog.Logger,
    e *internal.MonopolyEngine,
    playerId int,
    sessionId string,
    amount int,
    reason string,
) internal.UserActionStatus {
    if e.PendingBankPayout != nil {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "there is already a pending bank payout",
        }
    }

    if e.PendingBankPayment != nil {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "there is already a pending bank payment",
        }
    }

    e.PendingBankPayment = &internal.PendingBankPayment{
        PlayerId:  playerId,
        SessionId: sessionId,
        Amount:    amount,
        Reason:    reason,
    }

    e.Broker.Broadcast(log, "BankPaymentDueEvent", e.PendingBankPayment)

    return internal.UserActionStatus{
        Status: http.StatusOK,
        Data:   *e.PendingBankPayment,
    }
}

func SetPendingBankPayout(
    log zerolog.Logger,
    e *internal.MonopolyEngine,
    playerId int,
    sessionId string,
    amount int,
    reason string,
) internal.UserActionStatus {
    if e.PendingBankPayment != nil {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "there is already a pending bank payment",
        }
    }

    if e.PendingBankPayout != nil {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "there is already a pending bank payout",
        }
    }

    e.PendingBankPayout = &internal.PendingBankPayout{
        PlayerId:  playerId,
        SessionId: sessionId,
        Amount:    amount,
        Reason:    reason,
    }

    e.Broker.Broadcast(log, "BankPayoutDueEvent", e.PendingBankPayout)

    return internal.UserActionStatus{
        Status: http.StatusOK,
        Data:   *e.PendingBankPayout,
    }
}

func SetPendingPlayerExchange(
	log zerolog.Logger,
	e *internal.MonopolyEngine,
	actingPlayerId int,
	sessionId string,
	amount int,
	reason string,
	isPayingAll bool,
) internal.UserActionStatus {
	if e.PendingExchange != nil {
		return internal.UserActionStatus{
			Status: http.StatusBadRequest,
			Msg:    "there is already a pending player exchange",
		}
	}

	e.PendingExchange = &internal.PendingPlayerExchange{
		ActingPlayerId: actingPlayerId,
		SessionId:      sessionId,
		Amount:         amount,
		Reason:         reason,
		IsPayingAll:    isPayingAll,
	}

	e.Broker.Broadcast(log, "PlayerExchangeDueEvent", e.PendingExchange)

	return internal.UserActionStatus{
		Status: http.StatusOK,
		Data:   *e.PendingExchange,
	}
}

func clearPlayerTurnState(e *internal.MonopolyEngine, playerId int) {
    delete(e.PendingRolls, playerId)
    delete(e.TurnHasRolled, playerId)
    delete(e.ExtraRollAllowed, playerId)
    delete(e.DoubleRollCounts, playerId)
}

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

// getNewPosition returns the player's new board position after moving by the
// given total and reports whether the move passed Go.
// The board is treated as a 40-tile loop.
func getNewPosition(position int, total int) (int, bool) {
    newPosition := (position + total) % 40
    return newPosition, position+total >= 40
}

// Connected handles a player joining the running game session.
// It verifies that the player exists in the session, marks them as in-game,
// notifies connected clients, and sends the full initial board state.
func Connected(
    ctx context.Context,
    log zerolog.Logger,
    e *internal.MonopolyEngine,
    action *internal.UserActionEvent,
    tx *pgxpool.Tx,
) internal.UserActionStatus {
    log.Trace().Msg("player attempting to join game")

    data := action.Data.(struct {
        Id         int
        PlayerName string
        SessionId  string
    })

    // ensure player exists in session
    player_exists, err := internaldb_players.CheckPlayerExists(
        log,
        ctx,
        tx,
        data.Id,
        data.PlayerName,
        data.SessionId,
    )
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    if !player_exists {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "player does not exist",
        }
    }

    internaldb_players.UpdatePlayerInGameStatus(log, ctx, tx, data.Id, data.SessionId, true)

    // announce to all connected users that another user has joined the game
    e.Broker.BroadcastComment(log, fmt.Sprintf("Player %v has joined", data.PlayerName))

    err = events.EmitInitialGameBoardData(log, ctx, e, tx, data)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    log.Trace().Msgf("player %v has joined server", data.PlayerName)
    return internal.UserActionStatus{
        Status: http.StatusOK,
    }
}

// Disconnected handles a player leaving the running game session.
// It verifies that the player exists, marks them as out of the game, updates
// ready status when needed, notifies connected clients, and broadcasts a board
// state update.
func Disconnected(
    ctx context.Context,
    log zerolog.Logger,
    e *internal.MonopolyEngine,
    action *internal.UserActionEvent,
    tx *pgxpool.Tx,
) internal.UserActionStatus {

    log.Trace().Msg("player attempting to leave game")

    data := action.Data.(struct {
        Id         int
        PlayerName string
        SessionId  string
    })

    // ensure player exists in session
    player_exists, err := internaldb_players.CheckPlayerExists(
        log,
        ctx,
        tx,
        data.Id,
        data.PlayerName,
        data.SessionId,
    )
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    if !player_exists {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "player does not exist",
        }
    }

    internaldb_players.UpdatePlayerInGameStatus(log, ctx, tx, data.Id, data.SessionId, false)

    // check if all players are ready
    ready_stats, err := internaldb_players.GetAllPlayersReadyUpStatus(
        log,
        ctx,
        tx,
        data.SessionId,
    )
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Data:   err.Error(),
        }
    }

    if ready_stats.Ready != ready_stats.Total {
        // everyone is ready, don't reset ready up
        internaldb_players.SetPlayerReadyUpStatus(log, ctx, tx, data.Id, data.SessionId, false)
    }

    // announce to all connected users that another user has left the game
    e.Broker.BroadcastComment(log, fmt.Sprintf("Player %v has left", data.PlayerName))

    // only emit simple board update since you don't need to re-emit all the tile
    // info on disconnect
    err = events.EmitGameBoardUpdate(log, ctx, e, tx)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    log.Trace().Msgf("player %v has left server", data.PlayerName)
    return internal.UserActionStatus{
        Status: http.StatusOK,
    }
}

// ReadyUp updates a player's ready-up status for the session.
// After updating the player's state, it broadcasts a board update and starts
// the game when all players in the session are ready.
func ReadyUp(
    ctx context.Context,
    log zerolog.Logger,
    e *internal.MonopolyEngine,
    action *internal.UserActionEvent,
    tx *pgxpool.Tx,
) internal.UserActionStatus {
    data := action.Data.(struct {
        SessionId string
        PlayerId  int
        Status    bool
    })

    err := internaldb_players.SetPlayerReadyUpStatus(
        log,
        ctx,
        tx,
        data.PlayerId,
        data.SessionId,
        data.Status,
    )
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    events.EmitGameBoardUpdate(log, ctx, e, tx)

    // check if all players are ready
    ready_stats, err := internaldb_players.GetAllPlayersReadyUpStatus(
        log,
        ctx,
        tx,
        data.SessionId,
    )
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Data:   err.Error(),
        }
    }

    if ready_stats.Ready == ready_stats.Total {
        // everyone is ready
        internaldb_game_state.UpdateGameStateTurnNumber(log, ctx, tx, data.SessionId, 0)
        log.Info().Msg("final player has readied up; Start Game")
        e.Broker.Broadcast(log, "GameReadyEvent", "START")
    }

    return internal.UserActionStatus{
        Status: http.StatusOK,
        Data:   "",
    }
}

// RollDice rolls the dice for the current player and stores the result in the
// engine state.
// During turn-order setup, the roll is added to the temporary turn-decision
// list. During normal gameplay, the roll becomes a pending move for that
// player. The result is then broadcast to connected clients.
func RollDice(
    ctx context.Context,
    log zerolog.Logger,
    e *internal.MonopolyEngine,
    action *internal.UserActionEvent,
    tx *pgxpool.Tx,
) internal.UserActionStatus {
    data := action.Data.(internal.SimpleActionData)

    _, player, players, _, status := getActionPlayers(ctx, log, tx, data.SessionId, data.PlayerId)
    if status != nil {
        return *status
    }

    if e.TurnNumber >= len(players) && e.TurnHasRolled[data.PlayerId] && !e.ExtraRollAllowed[data.PlayerId] {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "player has already rolled this turn",
        }
    }

    diceRoll := internal.DiceRoll{
        PlayerId:  data.PlayerId,
        SessionId: data.SessionId,
        DieOne:    rand.IntN(6) + 1,
        DieTwo:    rand.IntN(6) + 1,
    }

	diceRoll = internal.DiceRoll{
        PlayerId:  data.PlayerId,
        SessionId: data.SessionId,
        DieOne:    1,
        DieTwo:    1,
    }
	
    diceRoll.Total = diceRoll.DieOne + diceRoll.DieTwo
    diceRoll.IsDouble = diceRoll.DieOne == diceRoll.DieTwo

    if e.TurnNumber >= 0 && e.TurnNumber < len(players) {
        rolls := e.TempStore["turn_decision_rolls"].([]internal.DiceRoll)
        rolls = append(rolls, diceRoll)

        // sort rolls so highest roll is first
        sort.SliceStable(rolls, func(i, j int) bool {
            if rolls[i].Total == rolls[j].Total {
                if rolls[i].DieOne == rolls[j].DieOne {
                    if rolls[i].DieTwo == rolls[j].DieTwo {
                        return rolls[i].PlayerId < rolls[j].PlayerId
                    }
                    return rolls[i].DieTwo > rolls[j].DieTwo
                }
                return rolls[i].DieOne > rolls[j].DieOne
            }
            return rolls[i].Total > rolls[j].Total
        })

        e.TempStore["turn_decision_rolls"] = rolls
    } else {
        if e.PendingRent != nil && e.PendingRent.FromPlayerId == data.PlayerId {
            return internal.UserActionStatus{
                Status: http.StatusBadRequest,
                Msg:    "player has a pending rent payment",
            }
        }

        if e.PendingBankPayment != nil && e.PendingBankPayment.PlayerId == data.PlayerId {
            return internal.UserActionStatus{
                Status: http.StatusBadRequest,
                Msg:    "player has a pending bank payment",
            }
        }

        if _, ok := e.PendingRolls[data.PlayerId]; ok {
            return internal.UserActionStatus{
                Status: http.StatusBadRequest,
                Msg:    "player already has a pending dice roll",
            }
        }

        e.TurnHasRolled[data.PlayerId] = true
        e.ExtraRollAllowed[data.PlayerId] = false

        if player.Jailed > 0 {
            if diceRoll.IsDouble {
                err := internaldb_players.UpdatePlayerJailState(
                    log,
                    ctx,
                    tx,
                    data.PlayerId,
                    data.SessionId,
                    player.GetOutOfJailCards,
                    0,
                )
                if err != nil {
                    return internal.UserActionStatus{
                        Status: http.StatusInternalServerError,
                        Msg:    err.Error(),
                    }
                }

                e.DoubleRollCounts[data.PlayerId] += 1
                diceRoll.ReleasedFromJail = true
                diceRoll.Jailed = 0
                e.PendingRolls[data.PlayerId] = diceRoll
            } else {
                jailedTurns := player.Jailed
                if jailedTurns < 3 {
                    jailedTurns += 1
                    err := internaldb_players.UpdatePlayerJailState(
                        log,
                        ctx,
                        tx,
                        data.PlayerId,
                        data.SessionId,
                        player.GetOutOfJailCards,
                        jailedTurns,
                    )
                    if err != nil {
                        return internal.UserActionStatus{
                            Status: http.StatusInternalServerError,
                            Msg:    err.Error(),
                        }
                    }
                }

                diceRoll.Jailed = jailedTurns
                e.DoubleRollCounts[data.PlayerId] = 0
                e.PendingRolls[data.PlayerId] = diceRoll
            }
        } else {
            if diceRoll.IsDouble {
                e.DoubleRollCounts[data.PlayerId] += 1
                if e.DoubleRollCounts[data.PlayerId] >= 3 {
                    err := internaldb_players.UpdatePlayerPositionAndJailed(log, ctx, tx, data.PlayerId, data.SessionId, 10, 1)
                    if err != nil {
                        return internal.UserActionStatus{
                            Status: http.StatusInternalServerError,
                            Msg:    err.Error(),
                        }
                    }

                    e.ExtraRollAllowed[data.PlayerId] = false
                    diceRoll.SentToJail = true
                    diceRoll.Jailed = 1
                } else {
                    diceRoll.RollAgain = true
                    e.PendingRolls[data.PlayerId] = diceRoll
                }
            } else {
                e.DoubleRollCounts[data.PlayerId] = 0
                e.PendingRolls[data.PlayerId] = diceRoll
            }
        }
    }

    e.Broker.Broadcast(log, "RollDiceEvent", diceRoll)
    events.EmitGameBoardUpdate(log, ctx, e, tx)

    return internal.UserActionStatus{
        Status: http.StatusOK,
        Data:   diceRoll,
    }
}

// EndTurn advances the game to the next turn for the current player.
// While the game is still deciding player order, it also assigns final turn
// order once all setup rolls have been completed. After updating turn state,
// it broadcasts the latest board update.
func EndTurn(
    ctx context.Context,
    log zerolog.Logger,
    e *internal.MonopolyEngine,
    action *internal.UserActionEvent,
    tx *pgxpool.Tx,
) internal.UserActionStatus {

    data := action.Data.(internal.SimpleActionData)

    _, _, players, _, status := getActionPlayers(ctx, log, tx, data.SessionId, data.PlayerId)
    if status != nil {
        return *status
    }

    if e.PendingBankPayment != nil && e.PendingBankPayment.PlayerId == data.PlayerId {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "player has a pending bank payment",
        }
    }

    if e.PendingBankPayout != nil && e.PendingBankPayout.PlayerId == data.PlayerId {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "player has a pending bank payout",
        }
    }

    if e.PendingRent != nil && e.PendingRent.FromPlayerId == data.PlayerId {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "player has a pending rent payment",
        }
    }

    if _, ok := e.PendingRolls[data.PlayerId]; ok {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "player has a pending dice roll",
        }
    }

    if e.TurnNumber >= len(players) {
        if !e.TurnHasRolled[data.PlayerId] {
            return internal.UserActionStatus{
                Status: http.StatusBadRequest,
                Msg:    "player must roll to start the turn",
            }
        }

        if e.ExtraRollAllowed[data.PlayerId] {
            return internal.UserActionStatus{
                Status: http.StatusBadRequest,
                Msg:    "player must roll again",
            }
        }
    }

    err := internaldb_game_state.UpdateGameStateTurnNumber(log, ctx, tx, data.SessionId, e.TurnNumber+1)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }
    // turn update succeeded, update internal counter
    e.TurnNumber += 1

    // all players have rolled to determine their turn order
    if e.TurnNumber == len(players) {
        // give all players their turn order
        for i, roll := range e.TempStore["turn_decision_rolls"].([]internal.DiceRoll) {
            log.Trace().Msgf("%#v", roll)
            err := internaldb_players.UpdatePlayerTurnOrder(
                log,
                ctx,
                tx,
                roll.PlayerId,
                roll.SessionId,
                i,
            )
            if err != nil {
                return internal.UserActionStatus{
                    Status: http.StatusInternalServerError,
                    Msg:    err.Error(),
                }
            }
        }
    }

    // FIXME: Prevent ending turn if there is a pending rent. If player can't pay
    // rent and is bankrupt, handle that here.

    clearPlayerTurnState(e, data.PlayerId)

    events.EmitGameBoardUpdate(log, ctx, e, tx)

    return internal.UserActionStatus{
        Status: http.StatusOK,
    }
}

// MovePlayer moves the current player using their pending dice roll.
// It validates that the player has already rolled, updates their board
// position, clears the pending roll, checks whether rent is now owed, and
// broadcasts either the movement alone or both the movement and a rent-due
// event.
func MovePlayer(
    ctx context.Context,
    log zerolog.Logger,
    e *internal.MonopolyEngine,
    action *internal.UserActionEvent,
    tx *pgxpool.Tx,
) internal.UserActionStatus {
    data := action.Data.(internal.SimpleActionData)

    _, _, _, turnNumber, status := getActionPlayers(ctx, log, tx, data.SessionId, data.PlayerId)
    if status != nil {
        return *status
    }

    if e.PendingBankPayment != nil && e.PendingBankPayment.PlayerId == data.PlayerId {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "player has a pending bank payment",
        }
    }

    diceRoll, ok := e.PendingRolls[data.PlayerId]
    if !ok {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "player does not have a pending dice roll",
        }
    }

    player, err := internaldb_players.GetPlayer(log, ctx, tx, data.PlayerId, data.SessionId)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    if player.Jailed > 0 {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "player is in jail",
        }
    }

    newPosition, passedGo := getNewPosition(player.Position, diceRoll.Total)
    err = internaldb_players.UpdatePlayerPosition(log, ctx, tx, data.PlayerId, data.SessionId, newPosition)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    delete(e.PendingRolls, data.PlayerId)

    playerMovement := internal.PlayerMovement{
        PlayerId:    data.PlayerId,
        SessionId:   data.SessionId,
        OldPosition: player.Position,
        NewPosition: newPosition,
        Total:       diceRoll.Total,
        PassedGo:    passedGo,
        TurnNumber:  turnNumber,
    }

    if diceRoll.IsDouble {
        e.ExtraRollAllowed[data.PlayerId] = true
        playerMovement.RollAgain = true
    } else {
        e.ExtraRollAllowed[data.PlayerId] = false
        e.DoubleRollCounts[data.PlayerId] = 0
    }

    tileData, err := internaldb_tiles.GetRentTileData(log, ctx, tx, data.SessionId, newPosition)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    playerMovement.PropertyId = tileData.PropertyId

	if tileData.Name == "Community Chest" || tileData.Name == "Chance" {
		var dbCardType string
		if tileData.Name == "Chance" {
			dbCardType = "CHANCE"
		} else {
			dbCardType = "COMMUNITY"
		}

		cardId, err := internaldb_event_cards.AssignEventCardDB(log, ctx, tx, data.SessionId, dbCardType, data.PlayerId)
		if err != nil {
			return internal.UserActionStatus{
				Status: http.StatusInternalServerError,
				Msg:    err.Error(),
			}
		}

		cardId = 24

		e.Broker.Broadcast(log, "DrawCardEvent", map[string]interface{}{
			"player_id":  data.PlayerId,
			"session_id": data.SessionId,
			"card_id":    cardId,
		})

		effectFunc, exists := CardEffects[cardId]
		if exists {
			status := effectFunc(ctx, log, e, tx, data.SessionId, data.PlayerId)
			if status.Status != http.StatusOK {
				return status
			}
		}

		updatedPlayer, err := internaldb_players.GetPlayer(log, ctx, tx, data.PlayerId, data.SessionId)
		if err != nil {
			return internal.UserActionStatus{Status: http.StatusInternalServerError, Msg: err.Error()}
		}

		// this is so that the tile data is updated if we move
		if updatedPlayer.Position != newPosition {
			newPosition = updatedPlayer.Position
			playerMovement.NewPosition = newPosition

			tileData, err = internaldb_tiles.GetRentTileData(log, ctx, tx, data.SessionId, newPosition)
			if err != nil {
				return internal.UserActionStatus{Status: http.StatusInternalServerError, Msg: err.Error()}
			}
			playerMovement.PropertyId = tileData.PropertyId
		}
	}

    if tileData.HasProperty && tileData.Owned && tileData.OwnerId != data.PlayerId && !tileData.IsMortgaged {
        isUtilityCard, _ := e.TempStore["special_utility_rent"].(bool)
        isRailroadCard, _ := e.TempStore["special_railroad_rent"].(bool)

        rentAmount, err := getRentAmount(ctx, log, tx, data.SessionId, tileData, diceRoll.Total, isUtilityCard, isRailroadCard)
        if err != nil {
            return internal.UserActionStatus{
                Status: http.StatusInternalServerError,
                Msg:    err.Error(),
            }
        }

        delete(e.TempStore, "special_utility_rent")
        delete(e.TempStore, "special_railroad_rent")

        if rentAmount > 0 {
            e.PendingRent = &internal.PendingRent{
                FromPlayerId: data.PlayerId,
                ToPlayerId:   tileData.OwnerId,
                SessionId:    data.SessionId,
                PropertyId:   tileData.PropertyId,
                Position:     newPosition,
                Amount:       rentAmount,
                DiceTotal:    diceRoll.Total,
                IsUtilityCard: isUtilityCard,
                IsRailroadCard: isRailroadCard,
            }

            playerMovement.RentDue = true
            playerMovement.RentAmount = rentAmount
            playerMovement.RentToId = tileData.OwnerId
            playerMovement.PropertyId = tileData.PropertyId

            e.Broker.Broadcast(log, "MovePlayerEvent", playerMovement)
            e.Broker.Broadcast(log, "RentDueEvent", e.PendingRent)

            return internal.UserActionStatus{
                Status: http.StatusOK,
                Data:   playerMovement,
            }
        }
    }

    e.Broker.Broadcast(log, "MovePlayerEvent", playerMovement)
    events.EmitGameBoardUpdate(log, ctx, e, tx)

    return internal.UserActionStatus{
        Status: http.StatusOK,
        Data:   playerMovement,
    }
}

func PayBank(
    ctx context.Context,
    log zerolog.Logger,
    e *internal.MonopolyEngine,
    action *internal.UserActionEvent,
    tx *pgxpool.Tx,
) internal.UserActionStatus {
    data := action.Data.(internal.BankPaymentActionData)

    if e.PendingBankPayment == nil {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "there is no pending bank payment",
        }
    }

    _, _, _, _, status := getActionPlayers(ctx, log, tx, data.SessionId, data.PlayerId)
    if status != nil {
        return *status
    }

    if e.PendingBankPayment.PlayerId != data.PlayerId || e.PendingBankPayment.SessionId != data.SessionId {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "bank payment payer is incorrect",
        }
    }

    player, err := internaldb_players.GetPlayer(log, ctx, tx, data.PlayerId, data.SessionId)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    if player.Money < e.PendingBankPayment.Amount {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "player does not have enough money",
        }
    }

    err = internaldb_players.UpdatePlayerMoney(
        log,
        ctx,
        tx,
        data.PlayerId,
        data.SessionId,
        player.Money-e.PendingBankPayment.Amount,
    )
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    bankPayment := internal.BankPayment{
        PlayerId:    data.PlayerId,
        SessionId:   data.SessionId,
        Amount:      e.PendingBankPayment.Amount,
        Reason:      e.PendingBankPayment.Reason,
        PlayerMoney: player.Money - e.PendingBankPayment.Amount,
        Jailed:      player.Jailed,
    }

    if e.PendingBankPayment.Reason == "jail release" {
        err = internaldb_players.UpdatePlayerJailState(
            log,
            ctx,
            tx,
            data.PlayerId,
            data.SessionId,
            player.GetOutOfJailCards,
            0,
        )
        if err != nil {
            return internal.UserActionStatus{
                Status: http.StatusInternalServerError,
                Msg:    err.Error(),
            }
        }

        bankPayment.Jailed = 0
        e.Broker.Broadcast(log, "PayToLeaveJailEvent", internal.JailRelease{
            PlayerId:          data.PlayerId,
            SessionId:         data.SessionId,
            Method:            "pay",
            GetOutOfJailCards: player.GetOutOfJailCards,
            PlayerMoney:       bankPayment.PlayerMoney,
            Jailed:            0,
        })
    }

	e.PendingBankPayment = nil

    e.Broker.Broadcast(log, "BankPaymentEvent", bankPayment)
    events.EmitGameBoardUpdate(log, ctx, e, tx)

    return internal.UserActionStatus{
        Status: http.StatusOK,
        Data:   bankPayment,
    }
}

func SetBankPayout(
    ctx context.Context,
    log zerolog.Logger,
    e *internal.MonopolyEngine,
    action *internal.UserActionEvent,
    tx *pgxpool.Tx,
) internal.UserActionStatus {
    data := action.Data.(internal.BankPayoutActionData)

    _, _, _, _, status := getActionPlayers(ctx, log, tx, data.SessionId, data.PlayerId)
    if status != nil {
        return *status
    }

    if data.Amount < 1 {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "amount must be greater than 0",
        }
    }

    return SetPendingBankPayout(log, e, data.PlayerId, data.SessionId, data.Amount, data.Reason)
}

func ReceiveBankPayout(
    ctx context.Context,
    log zerolog.Logger,
    e *internal.MonopolyEngine,
    action *internal.UserActionEvent,
    tx *pgxpool.Tx,
) internal.UserActionStatus {
    data := action.Data.(internal.BankPaymentActionData)

    if e.PendingBankPayout == nil {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "there is no pending bank payout",
        }
    }

    _, _, _, _, status := getActionPlayers(ctx, log, tx, data.SessionId, data.PlayerId)
    if status != nil {
        return *status
    }

    if e.PendingBankPayout.PlayerId != data.PlayerId || e.PendingBankPayout.SessionId != data.SessionId {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "bank payout recipient is incorrect",
        }
    }

    player, err := internaldb_players.GetPlayer(log, ctx, tx, data.PlayerId, data.SessionId)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    err = internaldb_players.UpdatePlayerMoney(
        log,
        ctx,
        tx,
        data.PlayerId,
        data.SessionId,
        player.Money+e.PendingBankPayout.Amount,
    )
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    bankPayout := internal.BankPayout{
        PlayerId:    data.PlayerId,
        SessionId:   data.SessionId,
        Amount:      e.PendingBankPayout.Amount,
        Reason:      e.PendingBankPayout.Reason,
        PlayerMoney: player.Money + e.PendingBankPayout.Amount,
    }

    e.PendingBankPayout = nil

    e.Broker.Broadcast(log, "BankPayoutEvent", bankPayout)
    events.EmitGameBoardUpdate(log, ctx, e, tx)

    return internal.UserActionStatus{
        Status: http.StatusOK,
        Data:   bankPayout,
    }
}

func ExecutePlayerExchange(
	ctx context.Context,
	log zerolog.Logger,
	e *internal.MonopolyEngine,
	action *internal.UserActionEvent,
	tx *pgxpool.Tx,
) internal.UserActionStatus {
	data := action.Data.(internal.PlayerExchangeActionData)

	if e.PendingExchange == nil {
		return internal.UserActionStatus{
			Status: http.StatusBadRequest,
			Msg:    "there is no pending player exchange",
		}
	}

	_, _, _, _, status := getActionPlayers(ctx, log, tx, data.SessionId, data.PlayerId)
	if status != nil {
		return *status
	}

	if e.PendingExchange.ActingPlayerId != data.PlayerId || e.PendingExchange.SessionId != data.SessionId {
		return internal.UserActionStatus{
			Status: http.StatusBadRequest,
			Msg:    "acting player for exchange is incorrect",
		}
	}

	players, err := internaldb_players.GetPlayersInSession(log, ctx, tx, data.SessionId)
	if err != nil {
		return internal.UserActionStatus{Status: http.StatusInternalServerError, Msg: err.Error()}
	}

	balances := make(map[int]int)
	actingPlayer, err := internaldb_players.GetPlayer(log, ctx, tx, data.PlayerId, data.SessionId)
	if err != nil {
		return internal.UserActionStatus{Status: http.StatusInternalServerError, Msg: err.Error()}
	}

	totalActingAdjustment := 0
	for _, p := range players {
		if p.Id == data.PlayerId {
			continue
		}

		if e.PendingExchange.IsPayingAll {
			totalActingAdjustment -= e.PendingExchange.Amount
			err = internaldb_players.UpdatePlayerMoney(log, ctx, tx, p.Id, data.SessionId, p.Money+e.PendingExchange.Amount)
			balances[p.Id] = p.Money + e.PendingExchange.Amount
		} else {
			totalActingAdjustment += e.PendingExchange.Amount
			err = internaldb_players.UpdatePlayerMoney(log, ctx, tx, p.Id, data.SessionId, p.Money-e.PendingExchange.Amount)
			balances[p.Id] = p.Money - e.PendingExchange.Amount
		}

		if err != nil {
			return internal.UserActionStatus{Status: http.StatusInternalServerError, Msg: err.Error()}
		}
	}

	err = internaldb_players.UpdatePlayerMoney(log, ctx, tx, data.PlayerId, data.SessionId, actingPlayer.Money+totalActingAdjustment)
	if err != nil {
		return internal.UserActionStatus{Status: http.StatusInternalServerError, Msg: err.Error()}
	}
	balances[data.PlayerId] = actingPlayer.Money + totalActingAdjustment

	exchange := internal.PlayerExchange{
		ActingPlayerId: data.PlayerId,
		SessionId:      data.SessionId,
		Amount:         e.PendingExchange.Amount,
		Reason:         e.PendingExchange.Reason,
		IsPayingAll:    e.PendingExchange.IsPayingAll,
		Balances:       balances,
	}

	e.PendingExchange = nil
	e.Broker.Broadcast(log, "PlayerExchangeEvent", exchange)
	events.EmitGameBoardUpdate(log, ctx, e, tx)

	return internal.UserActionStatus{
		Status: http.StatusOK,
		Data:   exchange,
	}
}

func Bankrupt(
    ctx context.Context,
    log zerolog.Logger,
    e *internal.MonopolyEngine,
    action *internal.UserActionEvent,
    tx *pgxpool.Tx,
) internal.UserActionStatus {
    data := action.Data.(internal.SimpleActionData)

    _, player, players, _, status := getActionPlayers(ctx, log, tx, data.SessionId, data.PlayerId)
    if status != nil {
        return *status
    }

    if e.PendingRent == nil && e.PendingBankPayment == nil {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "player does not have a pending payment obligation",
        }
    }

    bankruptcyRank, err := internaldb_players.GetNextBankruptcyRank(log, ctx, tx, data.SessionId)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    bankruptcy := internal.Bankruptcy{
        PlayerId:  data.PlayerId,
        SessionId: data.SessionId,
        Rank:      bankruptcyRank,
    }

    if e.PendingRent != nil && e.PendingRent.FromPlayerId == data.PlayerId {
        recipient, err := internaldb_players.GetPlayer(log, ctx, tx, e.PendingRent.ToPlayerId, data.SessionId)
        if err != nil {
            return internal.UserActionStatus{
                Status: http.StatusInternalServerError,
                Msg:    err.Error(),
            }
        }

        err = internaldb_tiles.TransferOwnedProperties(log, ctx, tx, data.SessionId, data.PlayerId, recipient.Id)
        if err != nil {
            return internal.UserActionStatus{
                Status: http.StatusInternalServerError,
                Msg:    err.Error(),
            }
        }

        err = internaldb_players.UpdatePlayerMoney(log, ctx, tx, recipient.Id, data.SessionId, recipient.Money+player.Money)
        if err != nil {
            return internal.UserActionStatus{
                Status: http.StatusInternalServerError,
                Msg:    err.Error(),
            }
        }

        bankruptcy.OwesRent = true
        bankruptcy.RentToId = recipient.Id
        e.PendingRent = nil
    } else if e.PendingBankPayment != nil && e.PendingBankPayment.PlayerId == data.PlayerId {
        err = internaldb_tiles.ReleaseOwnedPropertiesToBank(log, ctx, tx, data.SessionId, data.PlayerId)
        if err != nil {
            return internal.UserActionStatus{
                Status: http.StatusInternalServerError,
                Msg:    err.Error(),
            }
        }

        e.PendingBankPayment = nil
    } else {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "player does not have a pending payment obligation",
        }
    }

    err = internaldb_players.UpdatePlayerBankruptcy(log, ctx, tx, data.PlayerId, data.SessionId, true, bankruptcyRank, 0)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    clearPlayerTurnState(e, data.PlayerId)

    err = internaldb_game_state.UpdateGameStateTurnNumber(log, ctx, tx, data.SessionId, e.TurnNumber+1)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }
    e.TurnNumber += 1

    nonBankruptPlayerCount, err := internaldb_players.GetNonBankruptPlayerCount(log, ctx, tx, data.SessionId)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    if nonBankruptPlayerCount == 1 {
        winner, err := internaldb_players.GetRemainingNonBankruptPlayer(log, ctx, tx, data.SessionId)
        if err != nil {
            return internal.UserActionStatus{
                Status: http.StatusInternalServerError,
                Msg:    err.Error(),
            }
        }

        if winner.Rank == 0 {
            err = internaldb_players.UpdatePlayerRank(log, ctx, tx, winner.Id, data.SessionId, 1)
            if err != nil {
                return internal.UserActionStatus{
                    Status: http.StatusInternalServerError,
                    Msg:    err.Error(),
                }
            }
        }

        bankruptcy.WinnerId = winner.Id
    }

    if e.TurnNumber == len(players) {
        for i, roll := range e.TempStore["turn_decision_rolls"].([]internal.DiceRoll) {
            err := internaldb_players.UpdatePlayerTurnOrder(
                log,
                ctx,
                tx,
                roll.PlayerId,
                roll.SessionId,
                i,
            )
            if err != nil {
                return internal.UserActionStatus{
                    Status: http.StatusInternalServerError,
                    Msg:    err.Error(),
                }
            }
        }
    }

    e.Broker.Broadcast(log, "BankruptcyEvent", bankruptcy)
    events.EmitGameBoardUpdate(log, ctx, e, tx)

    return internal.UserActionStatus{
        Status: http.StatusOK,
        Data:   bankruptcy,
    }
}

func ReleaseFromJail(
    ctx context.Context,
    log zerolog.Logger,
    e *internal.MonopolyEngine,
    action *internal.UserActionEvent,
    tx *pgxpool.Tx,
) internal.UserActionStatus {
    data := action.Data.(internal.JailReleaseActionData)

    _, player, _, _, status := getActionPlayers(ctx, log, tx, data.SessionId, data.PlayerId)
    if status != nil {
        return *status
    }

    if player.Jailed == 0 {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "player is not in jail",
        }
    }

    if e.PendingBankPayment != nil && e.PendingBankPayment.PlayerId == data.PlayerId {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "player has a pending bank payment",
        }
    }

    jailRelease := internal.JailRelease{
        PlayerId:          data.PlayerId,
        SessionId:         data.SessionId,
        Method:            data.Method,
        GetOutOfJailCards: player.GetOutOfJailCards,
        PlayerMoney:       player.Money,
        Jailed:            0,
    }

    diceRoll, ok := e.PendingRolls[data.PlayerId]
    if !ok {
        if data.Method == "pay" {
            return internal.UserActionStatus{
                Status: http.StatusBadRequest,
                Msg:    "player must roll before paying to get out of jail",
            }
        }

        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "player must roll before using a get out of jail card",
        }
    }

    if data.Method == "pay" {
        if diceRoll.DieOne == diceRoll.DieTwo {
            return internal.UserActionStatus{
                Status: http.StatusBadRequest,
                Msg:    "player rolled doubles and does not need to pay to get out of jail",
            }
        }

        if player.Money < 50 {
            pendingBankPaymentStatus := SetPendingBankPayment(log, e, data.PlayerId, data.SessionId, 50, "jail release")
            if pendingBankPaymentStatus.Status != http.StatusOK {
                return pendingBankPaymentStatus
            }

            return internal.UserActionStatus{
                Status: http.StatusOK,
                Data:   pendingBankPaymentStatus.Data,
            }
        }

        pendingBankPaymentStatus := SetPendingBankPayment(log, e, data.PlayerId, data.SessionId, 50, "jail release")
        if pendingBankPaymentStatus.Status != http.StatusOK {
            return pendingBankPaymentStatus
        }

        return internal.UserActionStatus{
            Status: http.StatusOK,
            Data:   pendingBankPaymentStatus.Data,
        }
    }

    if player.GetOutOfJailCards < 1 {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "player does not have a get out of jail card",
        }
    }

    if diceRoll.DieOne == diceRoll.DieTwo {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "player rolled doubles and does not need to use a get out of jail card",
        }
    }

    err := internaldb_players.UpdatePlayerJailState(
        log,
        ctx,
        tx,
        data.PlayerId,
        data.SessionId,
        player.GetOutOfJailCards-1,
        0,
    )
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    jailRelease.GetOutOfJailCards = player.GetOutOfJailCards - 1

    e.Broker.Broadcast(log, "UseGetOutOfJailCardEvent", jailRelease)
    events.EmitGameBoardUpdate(log, ctx, e, tx)

    return internal.UserActionStatus{
        Status: http.StatusOK,
        Data:   jailRelease,
    }
}
