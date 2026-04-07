package monopoly_engine

import (
	"context"
	"fmt"
	"monopoly-backend/handlers"
	"monopoly-backend/internal"
	internaldb "monopoly-backend/internal/db"
	internaldb_players "monopoly-backend/internal/db/players"
	internaldb_properties "monopoly-backend/internal/db/properties"
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
    action_status = <- actionEvent.ReturnChan

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
            Msg: err.Error(),
        }
        return nil
    }

	defer tx.Rollback(ctx)

    switch action.Event {
    case "ConnectionEvent":
        log.Trace().Msg("player attempting to join game")

        data := action.Data.(struct {
            Id string
            PlayerName string
            SessionId string
        })

        // ensure player exists in session
        player_exists, err := internaldb_players.CheckPlayerExists(
            log,
            ctx,
            tx.(*pgxpool.Tx),
            data.Id,
            data.PlayerName,
            data.SessionId,
            )
        if err != nil {
            action.ReturnChan <- internal.UserActionStatus{
                Status: http.StatusInternalServerError,
                Msg: err.Error(),
            }
            
			return nil
        }

        if !player_exists {
            action.ReturnChan <- internal.UserActionStatus{
                Status: http.StatusBadRequest,
                Msg: "player does not exist",
            }
            return nil
        }

        // announce to all connected users that another user has joined the game
        e.Broker.Broadcast(log, "ConnectionEvent", fmt.Sprintf("player %v has joined", data.PlayerName))
        action.ReturnChan <- internal.UserActionStatus{
            Status: http.StatusOK,
        }
        log.Trace().Msgf("player %v has joined server", data.PlayerName)

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
        action.ReturnChan <- internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg: "unknown action", 
        }
		return nil
    }

    err = tx.Commit(ctx)
    if err != nil {
        action.ReturnChan <- internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg: err.Error(),
        }
        return nil
    }
    return nil
}
