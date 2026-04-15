package monopoly_engine

import (
	"context"
	"fmt"
	"math/rand/v2"
	"monopoly-backend/handlers"
	"monopoly-backend/internal"
	internaldb "monopoly-backend/internal/db"
	internaldb_game_state "monopoly-backend/internal/db/game_state"
	internaldb_players "monopoly-backend/internal/db/players"
	internaldb_properties "monopoly-backend/internal/db/properties"
	"monopoly-backend/internal/engine/events/player_events"
	"net/http"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
    engineManager   = map[string]*internal.MonopolyEngine{}
    engineManagerMu sync.Mutex
)

func GetEngineBroker(sessionId string) (*handlers.SseBroker, error) {
    engine := engineManager[sessionId]
    if engine == nil {
        return nil, fmt.Errorf("engine is nil for provided session_id")
    }
    broker := engine.Broker
    if broker == nil {
        return nil, fmt.Errorf("engine exists but broker is nil. HUHH?")
    }
    return broker, nil
}

// NotifyEngineOfAction passes a user action event to the engine
func NotifyEngineOfAction(sessionId string, actionEvent internal.UserActionEvent) (internal.UserActionStatus, error) {
    var action_status internal.UserActionStatus
    engine, ok := engineManager[sessionId]
    if !ok {
        return action_status, fmt.Errorf("engine does not exist for provided session_id at this point in time")
    }

    engine.UserActionsChan <- actionEvent
    action_status = <-actionEvent.ReturnChan

    return action_status, nil
}

// SetupNewMonopolyEngine sets up new monopoly game with own unique state and players.
// This is typically run in a goroutine and lives for life of server.
func SetupNewMonopolyEngine(sessionId string) {
    log := log.Logger.With().Str("session_id", sessionId).Logger()

    engine := internal.MonopolyEngine{
        SessionId:       sessionId,
        Broker:          handlers.NewSseBroker(),
        UserActionsChan: make(chan internal.UserActionEvent),
        PendingRolls:    map[int]internal.DiceRoll{},
    }

    engineManagerMu.Lock()
    engineManager[sessionId] = &engine
    engineManagerMu.Unlock()

    ctx := context.Background()

    // infinite loop so we can recover from errors
    for {

        db, err := internaldb.CreateDbPoolConnection(ctx, log)
        if err != nil {
            log.Error().Err(err).Msg("failed to connect to database")
            return
        }

        err = runMonopolyEngine(ctx, log, &engine, db)
        if err == nil {
            return
        }
    }
}

func runMonopolyEngine(ctx context.Context, log zerolog.Logger, e *internal.MonopolyEngine, db *pgxpool.Pool) error {
    log.Info().Msgf("started monopoly engine for session id: %v", e.SessionId)

    for {

        select {
        case action, ok := <-e.UserActionsChan:
            if !ok {
                log.Warn().Msg("User Action channel was closed; exiting engine run loop")
                return fmt.Errorf("user action channel closed")
            }
            if err := processUserAction(ctx, log, e, action, db); err != nil {
                return err
            }
        case <-ctx.Done():
            log.Info().Msg("context was cancelled; engine stopped")
            return nil
        }

    }

}

func processUserAction(
    ctx context.Context,
    log zerolog.Logger,
    e *internal.MonopolyEngine,
    action internal.UserActionEvent,
    db *pgxpool.Pool,
) error {

    // NOTE: Here is where user actions will be handled
    log.Trace().Msgf("got action event: %v with data: %v", action.Event, action.Data)
    tx, err := db.BeginTx(ctx, pgx.TxOptions{})
    if err != nil {
        action.ReturnChan <- internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
        return nil
    }

    action_status := internal.UserActionStatus{
        Status: http.StatusOK,
    }

    switch action.Event {
    case "ConnectionEvent":
        action_status = player_events.Connected(ctx, log, e, &action, tx.(*pgxpool.Tx))
    case "DisconnectEvent":
        action_status = player_events.Disconnected(ctx, log, e, &action, tx.(*pgxpool.Tx))
    case "RollDiceEvent":
        data := action.Data.(internal.RollDiceActionData)

        players, err := internaldb_players.GetPlayersInSession(log, ctx, tx.(*pgxpool.Tx), data.SessionId)
        if err != nil {
            action_status = internal.UserActionStatus{
                Status: http.StatusInternalServerError,
                Msg:    err.Error(),
            }
            break
        }

        if len(players) == 0 {
            action_status = internal.UserActionStatus{
                Status: http.StatusBadRequest,
                Msg:    "there are no players in this game session",
            }
            break
        }

        turnNumber, err := internaldb_game_state.GetGameStateTurnNumber(log, ctx, tx.(*pgxpool.Tx), data.SessionId)
        if err != nil {
            action_status = internal.UserActionStatus{
                Status: http.StatusInternalServerError,
                Msg:    err.Error(),
            }
            break
        }

        currentPlayer := players[getCurrentPlayerIndex(turnNumber, len(players))]
        if currentPlayer.Id != data.PlayerId {
            action_status = internal.UserActionStatus{
                Status: http.StatusBadRequest,
                Msg:    "it is not this player's turn",
            }
            break
        }

        if _, ok := e.PendingRolls[data.PlayerId]; ok {
            action_status = internal.UserActionStatus{
                Status: http.StatusBadRequest,
                Msg:    "player already has a pending dice roll",
            }
            break
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

        action_status = internal.UserActionStatus{
            Status: http.StatusOK,
            Data:   diceRoll,
        }

    case "MovePlayerEvent":
        data := action.Data.(internal.MovePlayerActionData)

        players, err := internaldb_players.GetPlayersInSession(log, ctx, tx.(*pgxpool.Tx), data.SessionId)
        if err != nil {
            action_status = internal.UserActionStatus{
                Status: http.StatusInternalServerError,
                Msg:    err.Error(),
            }
            break
        }

        if len(players) == 0 {
            action_status = internal.UserActionStatus{
                Status: http.StatusBadRequest,
                Msg:    "there are no players in this game session",
            }
            break
        }

        turnNumber, err := internaldb_game_state.GetGameStateTurnNumber(log, ctx, tx.(*pgxpool.Tx), data.SessionId)
        if err != nil {
            action_status = internal.UserActionStatus{
                Status: http.StatusInternalServerError,
                Msg:    err.Error(),
            }
            break
        }

        currentPlayer := players[getCurrentPlayerIndex(turnNumber, len(players))]
        if currentPlayer.Id != data.PlayerId {
            action_status = internal.UserActionStatus{
                Status: http.StatusBadRequest,
                Msg:    "it is not this player's turn",
            }
            break
        }

        diceRoll, ok := e.PendingRolls[data.PlayerId]
        if !ok {
            action_status = internal.UserActionStatus{
                Status: http.StatusBadRequest,
                Msg:    "player does not have a pending dice roll",
            }
            break
        }

        player, err := internaldb_players.GetPlayer(log, ctx, tx.(*pgxpool.Tx), data.PlayerId, data.SessionId)
        if err != nil {
            action_status = internal.UserActionStatus{
                Status: http.StatusInternalServerError,
                Msg:    err.Error(),
            }
            break
        }

        newPosition, passedGo := getNewPosition(player.Position, diceRoll.Total)
        err = internaldb_players.UpdatePlayerPosition(log, ctx, tx.(*pgxpool.Tx), data.PlayerId, data.SessionId, newPosition)
        if err != nil {
            action_status = internal.UserActionStatus{
                Status: http.StatusInternalServerError,
                Msg:    err.Error(),
            }
            break
        }

        nextTurnNumber := getCurrentPlayerIndex(turnNumber, len(players)) + 1
        err = internaldb_game_state.UpdateGameStateTurnNumber(log, ctx, tx.(*pgxpool.Tx), data.SessionId, nextTurnNumber)
        if err != nil {
            action_status = internal.UserActionStatus{
                Status: http.StatusInternalServerError,
                Msg:    err.Error(),
            }
            break
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

        action_status = internal.UserActionStatus{
            Status: http.StatusOK,
            Data:   playerMovement,
        }

    case "PurchaseProperty":
        log.Trace().Msg("player attempting to purchase property")

        data, ok := action.Data.(struct {
            SessionId  string
            PlayerId   int
            PropertyId int
        })
        if !ok {
            log.Error().Msg("invalid data format received for PurchaseProperty")
            action.ReturnChan <- internal.UserActionStatus{
                Status: http.StatusBadRequest,
                Msg:    "internal data format error",
            }
            break
        }

        // try to buy property, also does ownership validation
        ownershipId, err := internaldb_properties.CreatePropertyOwnerDB(
            log,
            ctx,
            tx.(*pgxpool.Tx),
            data.SessionId,
            data.PlayerId,
            data.PropertyId,
        )

        if err != nil {
            action.ReturnChan <- internal.UserActionStatus{
                Status: http.StatusBadRequest,
                Msg:    err.Error(),
            }
            return nil
        }

        e.Broker.Broadcast(log, "PropertyPurchased", fmt.Sprintf("Player %d purchased property %d", data.PlayerId, data.PropertyId))

        action.ReturnChan <- internal.UserActionStatus{
            Status: http.StatusOK,
            Msg:    fmt.Sprintf("property purchased successfully (Ownership ID: %d)", ownershipId),
        }
        log.Trace().Msgf("player %d successfully purchased property %d", data.PlayerId, data.PropertyId)

        break

    default:
        log.Trace().Msgf("received unknown engine action event: %v", action.Event)
        action_status = internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    "unknown action",
        }
    }

    if action_status.Status != http.StatusOK {
        if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
            return rollbackErr
        }
        action.ReturnChan <- action_status
        return nil
    }

    err = tx.Commit(ctx)
    if err != nil {
        action.ReturnChan <- internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
        return nil
    }
    action.ReturnChan <- action_status
    return nil
}

func getCurrentPlayerIndex(turnNumber int, playerCount int) int {
    if turnNumber < 0 {
        return 0
    }

    return turnNumber % playerCount
}

func getNewPosition(position int, total int) (int, bool) {
    newPosition := (position + total) % 40
    return newPosition, position+total >= 40
}
