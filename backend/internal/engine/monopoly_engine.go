package monopoly_engine

import (
	"context"
	"fmt"
	"monopoly-backend/handlers"
	"monopoly-backend/internal"
	internaldb "monopoly-backend/internal/db"
	"monopoly-backend/internal/engine/events/player_events"
	"monopoly-backend/internal/engine/events/property_events"
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
        action_status = player_events.RollDice(ctx, log, e, &action, tx.(*pgxpool.Tx))
    case "MovePlayerEvent":
        action_status = player_events.MovePlayer(ctx, log, e, &action, tx.(*pgxpool.Tx))
    case "PurchaseProperty":
        action_status = property_events.PurchaseProperty(ctx, log, e, &action, tx.(*pgxpool.Tx))
    case "PurchaseHouse":
        action_status = property_events.PurchaseHouse(ctx, log, e, &action, tx.(*pgxpool.Tx))
    case "PurchaseHotel":
        action_status = property_events.PurchaseHotel(ctx, log, e, &action, tx.(*pgxpool.Tx))
    case "SellHouse":
        action_status = property_events.SellHouse(ctx, log, e, &action, tx.(*pgxpool.Tx))
    case "SellHotel":
        action_status = property_events.SellHotel(ctx, log, e, &action, tx.(*pgxpool.Tx))
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
