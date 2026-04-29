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

// GetEngineBroker returns the SSE broker for the given game session.
// It returns an error if no engine exists for the session or if the broker
// has not been initialized.
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

// NotifyEngineOfAction sends a user action to the engine for the given session
// and waits for the action result to be returned.
// Use this when an incoming request needs to be processed by the running engine.
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

// SetupNewMonopolyEngine creates and starts a new engine for a session.
// It initializes in-memory engine state, stores the engine in the global
// manager, and keeps the engine running for the lifetime of the server.
//
// This function is typically started in a goroutine.
func SetupNewMonopolyEngine(sessionId string) {
    log := log.Logger.With().Str("session_id", sessionId).Logger()

    engine := internal.MonopolyEngine{
        SessionId:       sessionId,
        Broker:          handlers.NewSseBroker(),
        UserActionsChan: make(chan internal.UserActionEvent),
        TempStore:       make(map[string]any),
        PendingRolls:    map[int]internal.DiceRoll{},
        PendingRent:     nil,
        PendingBankPayment: nil,
        TurnHasRolled:   map[int]bool{},
        ExtraRollAllowed: map[int]bool{},
        DoubleRollCounts: map[int]int{},
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

// runMonopolyEngine initializes engine state from the database and then enters
// the main event loop for the session.
// It resets player connection state, restores the current turn state, handles
// partial turn-order setup after restarts, and continuously processes user
// actions until the context is cancelled or an error occurs.
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
        e.TempStore["turn_decision_rolls"] = make([]internal.DiceRoll, 0)
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

// processUserAction handles a single user action inside a database transaction.
// It routes the action to the correct handler, rolls back the transaction if
// the action fails, and commits the transaction if the action succeeds.
//
// This should be used by the engine loop so that each action is processed
// serially and returns a status to the caller through the action's return
// channel.
func processUserAction(
    ctx context.Context,
    log zerolog.Logger,
    e *internal.MonopolyEngine,
    action internal.UserActionEvent,
    db *pgxpool.Pool,
) error {

    //Here is where user actions are be handled
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
    case "EndTurnEvent":
        action_status = player.EndTurn(ctx, log, e, &action, tx.(*pgxpool.Tx))
    case "RollDiceEvent":
        action_status = player.RollDice(ctx, log, e, &action, tx.(*pgxpool.Tx))
    case "MovePlayerEvent":
        action_status = player.MovePlayer(ctx, log, e, &action, tx.(*pgxpool.Tx))
    case "ReleaseFromJailEvent":
        action_status = player.ReleaseFromJail(ctx, log, e, &action, tx.(*pgxpool.Tx))
    case "BankPaymentEvent":
        action_status = player.PayBank(ctx, log, e, &action, tx.(*pgxpool.Tx))
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
