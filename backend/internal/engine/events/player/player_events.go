package player

import (
    "context"
    "fmt"
    "math/rand/v2"
    "monopoly-backend/internal"
    internaldb_game_state "monopoly-backend/internal/db/game_state"
    internaldb_players "monopoly-backend/internal/db/player"
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
        Id         string
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
    //e.Broker.Broadcast(log, "ConnectionEvent", fmt.Sprintf("player %v has joined", data.PlayerName))
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
        Id         string
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

    // announce to all connected users that another user has left the game
    e.Broker.BroadcastComment(log, fmt.Sprintf("Player %v has left", data.PlayerName))
    //e.Broker.Broadcast(log, "DisconnectEvent", fmt.Sprintf("player %v has left", data.PlayerName))
    log.Trace().Msgf("player %v has left server", data.PlayerName)

    return internal.UserActionStatus{
        Status: http.StatusOK,
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

    currentPlayer, players, turnNumber, err := turn_events.GetCurrentPlayer(ctx, log, tx, data.SessionId)
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

    nextTurnNumber := turn_events.GetCurrentPlayerIndex(turnNumber, len(players)) + 1
    err = internaldb_game_state.UpdateGameStateTurnNumber(log, ctx, tx, data.SessionId, nextTurnNumber)
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
        TurnNumber:  nextTurnNumber,
    }
    e.Broker.Broadcast(log, "MovePlayerEvent", playerMovement)

    return internal.UserActionStatus{
        Status: http.StatusOK,
        Data:   playerMovement,
    }
}
