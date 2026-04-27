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

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

func getNewPosition(position int, total int) (int, bool) {
    newPosition := (position + total) % 40
    return newPosition, position+total >= 40
}


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

func RollDice(
    ctx context.Context,
    log zerolog.Logger,
    e *internal.MonopolyEngine,
    action *internal.UserActionEvent,
    tx *pgxpool.Tx,
) internal.UserActionStatus {
    data := action.Data.(internal.RollDiceActionData)

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

    diceRoll := internal.DiceRoll{
        PlayerId:  data.PlayerId,
        SessionId: data.SessionId,
        DieOne:    rand.IntN(6) + 1,
        DieTwo:    rand.IntN(6) + 1,
    }
    diceRoll.Total = diceRoll.DieOne + diceRoll.DieTwo
    e.PendingRolls[data.PlayerId] = diceRoll

    e.Broker.Broadcast(log, "RollDiceEvent", diceRoll)
    events.EmitGameBoardUpdate(log, ctx, e, tx)

    return internal.UserActionStatus{
        Status: http.StatusOK,
        Data:   diceRoll,
    }
}

func MovePlayer(
    ctx context.Context,
    log zerolog.Logger,
    e *internal.MonopolyEngine,
    action *internal.UserActionEvent,
    tx *pgxpool.Tx,
) internal.UserActionStatus {
    data := action.Data.(internal.MovePlayerActionData)

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
        PlayerId:    data.PlayerId,
        SessionId:   data.SessionId,
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
