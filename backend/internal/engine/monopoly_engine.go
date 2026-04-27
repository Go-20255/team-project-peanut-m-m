package monopoly_engine

import (
	"context"
	"fmt"
	"monopoly-backend/handlers"
	"monopoly-backend/internal"
	internaldb "monopoly-backend/internal/db"
	internaldb_game_state "monopoly-backend/internal/db/game_state"
	internaldb_players "monopoly-backend/internal/db/player"
	"monopoly-backend/internal/engine/events/player"
	"monopoly-backend/internal/engine/events/property"
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
        TempStore:       make(map[string]any),
        PendingRolls:    map[int]internal.DiceRoll{},
        PendingRent:     nil,
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
    log.Info().Msg("started monopoly engine")

    // NOTE: About turn number as states:
    // Turn == -1: Remains -1 until all players have readied up
    // Turn == 0: All players are ready, start with first player who was created to roll first for deciding order
    // 0 <= Turn < len(players): Deciding player order
    // len(players) <= Turn: Turns have decided, start with player whose Turn == current_turn % len(players)


    tx, err := db.BeginTx(ctx, pgx.TxOptions{})
    if err != nil {
        return err
    }

    internaldb_players.ResetAllPlayersInGameStatus(log, ctx, tx.(*pgxpool.Tx), e.SessionId)

    log.Info().Msg("reset all player's in game status'")

    // get current turn number of db
    current_turn, err := internaldb_game_state.GetGameStateTurnNumber(log, ctx, tx.(*pgxpool.Tx), e.SessionId)
    if err != nil {
        return err
    }

    // set the current turn number in the engine
    e.TurnNumber = current_turn

    // check if all players are ready
    ready_stats, err := internaldb_players.GetAllPlayersReadyUpStatus(
        log,
        ctx,
        tx.(*pgxpool.Tx),
        e.SessionId,
        )
    if err != nil {
        return err
    }

    if ready_stats.Ready == ready_stats.Total && e.TurnNumber < 0 {
        // everyone is ready and turn number is still -1 for some reason
        internaldb_game_state.UpdateGameStateTurnNumber(log, ctx, tx.(*pgxpool.Tx), e.SessionId, 0)
        log.Info().Msg("all players are ready; Start Game")
        e.Broker.Broadcast(log, "GameReady", "START")
    }


    players, err := internaldb_players.GetPlayersInSession(log, ctx, tx.(*pgxpool.Tx), e.SessionId)
    if err != nil {
        return err
    }

    // if we were in the middle of deciding turn order after everyone readied up
    // (i.e. 0 <= turn number < len(players)), reset turns so everyone can re roll as
    // we have already lost everyone's previous rolls since engine restart
    if e.TurnNumber < len(players) && e.TurnNumber >= 0 {
        log.Trace().Msg("was in middle of deciding turn over; reset turn number to 0")
        internaldb_game_state.UpdateGameStateTurnNumber(log, ctx, tx.(*pgxpool.Tx), e.SessionId, 0)
        e.TurnNumber = 0
    }

    err = tx.Commit(ctx)
    if err != nil {
        if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
            return rollbackErr
        }
        return err
    }

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
        action_status = player.Connected(ctx, log, e, &action, tx.(*pgxpool.Tx))
    case "DisconnectEvent":
        action_status = player.Disconnected(ctx, log, e, &action, tx.(*pgxpool.Tx))
    case "PlayerReadyUpEvent":
        action_status = player.ReadyUp(ctx, log, e, &action, tx.(*pgxpool.Tx))
    case "RollDiceEvent":
        action_status = player.RollDice(ctx, log, e, &action, tx.(*pgxpool.Tx))
    case "MovePlayerEvent":
        action_status = player.MovePlayer(ctx, log, e, &action, tx.(*pgxpool.Tx))
    case "PayRentEvent":
        action_status = player.PayRent(ctx, log, e, &action, tx.(*pgxpool.Tx))
    case "PurchaseProperty":
        action_status = property.PurchaseProperty(ctx, log, e, &action, tx.(*pgxpool.Tx))
    case "PurchaseHouse":
        action_status = property.PurchaseHouse(ctx, log, e, &action, tx.(*pgxpool.Tx))
    case "PurchaseHotel":
        action_status = property.PurchaseHotel(ctx, log, e, &action, tx.(*pgxpool.Tx))
    case "SellHouse":
        action_status = property.SellHouse(ctx, log, e, &action, tx.(*pgxpool.Tx))
    case "SellHotel":
        action_status = property.SellHotel(ctx, log, e, &action, tx.(*pgxpool.Tx))
    case "MortgageProperty":
        action_status = property.MortgageProperty(ctx, log, e, &action, tx.(*pgxpool.Tx))
    case "UnmortgageProperty":
        action_status = property.UnmortgageProperty(ctx, log, e, &action, tx.(*pgxpool.Tx))
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
