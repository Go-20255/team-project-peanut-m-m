package player

import (
	"context"
	"fmt"
	"math/rand/v2"
	"monopoly-backend/internal"
	internaldb_game_state "monopoly-backend/internal/db/game_state"
	internaldb_players "monopoly-backend/internal/db/player"
	internaldb_tiles "monopoly-backend/internal/db/tile"
	"monopoly-backend/internal/engine/events"
	turn_events "monopoly-backend/internal/engine/events/turn"
	"net/http"
	"sort"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

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
            Data: err.Error(),
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
        SessionId   string
        PlayerId    int
        Status      bool
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
            Data: err.Error(),
        }
    }

    if ready_stats.Ready == ready_stats.Total {
        // everyone is ready
        internaldb_game_state.UpdateGameStateTurnNumber(log, ctx, tx,data.SessionId, 0)
        log.Info().Msg("final player has readied up; Start Game")
        e.Broker.Broadcast(log, "GameReady", "START")
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

    currentPlayer, players, _, err := turn_events.GetCurrentPlayer(ctx, log, tx, data.SessionId)
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

    // FIXME: If I roll, then move, and if I haven't ended my turn, I can roll again
    // and move again and keep repeating.

    diceRoll := internal.DiceRoll{
        PlayerId: data.PlayerId,
        SessionId: data.SessionId,
        DieOne:    rand.IntN(6) + 1,
        DieTwo:    rand.IntN(6) + 1,
    }
    diceRoll.Total = diceRoll.DieOne + diceRoll.DieTwo

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

        if _, ok := e.PendingRolls[data.PlayerId]; ok {
            return internal.UserActionStatus{
                Status: http.StatusBadRequest,
                Msg:    "player already has a pending dice roll",
            }
        }
        e.PendingRolls[data.PlayerId] = diceRoll
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

    currentPlayer, players, _, err := turn_events.GetCurrentPlayer(ctx, log, tx, data.SessionId)
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

    err = internaldb_game_state.UpdateGameStateTurnNumber(log, ctx, tx, data.SessionId, e.TurnNumber + 1)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg: err.Error(),
        }
    }
    // turn update succeeded, update internal counter
    e.TurnNumber += 1

    // all players have rolled to determine their turn order
    if e.TurnNumber == len(players) {
        // give all players their turn order
        for i, roll := range e.TempStore["turn_decision_rolls"].([]internal.DiceRoll) {
            log.Trace().Msgf("%#v",roll)
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
                    Msg: err.Error(),
                }
            }
        }
    }

    // TODO:: Prevent ending turn if there is a pending rent. If player can't pay
    // rent and is bankrupt, handle that here.

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

    currentPlayer, _, turnNumber, err := turn_events.GetCurrentPlayer(ctx, log, tx, data.SessionId)
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
        PlayerId: data.PlayerId,
        SessionId: data.SessionId,
        OldPosition: player.Position,
        NewPosition: newPosition,
        Total:       diceRoll.Total,
        PassedGo:    passedGo,
        TurnNumber:  turnNumber,
    }

    tileData, err := internaldb_tiles.GetRentTileData(log, ctx, tx, data.SessionId, newPosition)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    playerMovement.PropertyId = tileData.PropertyId

    if tileData.HasProperty && tileData.Owned && tileData.OwnerId != data.PlayerId && !tileData.IsMortgaged {
        rentAmount, err := getRentAmount(ctx, log, tx, data.SessionId, tileData, diceRoll.Total)
        if err != nil {
            return internal.UserActionStatus{
                Status: http.StatusInternalServerError,
                Msg:    err.Error(),
            }
        }

        if rentAmount > 0 {
            e.PendingRent = &internal.PendingRent{
                FromPlayerId: data.PlayerId,
                ToPlayerId:   tileData.OwnerId,
                SessionId:    data.SessionId,
                PropertyId:   tileData.PropertyId,
                Position:     newPosition,
                Amount:       rentAmount,
                DiceTotal:    diceRoll.Total,
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
