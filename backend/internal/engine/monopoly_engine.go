package monopoly_engine

import (
	"context"
	"fmt"
	"monopoly-backend/handlers"
	internaldb "monopoly-backend/internal/db"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
    engine_manager = map[string]*MonopolyEngine{}
    engine_manager_mu sync.Mutex
)

type UserActionEvent struct {
    Event string
    Data string
}

func GetEngineBroker(session_id string) (*handlers.SseBroker, error) {
    engine := engine_manager[session_id]
    if engine == nil {
        return nil, fmt.Errorf("engine is nil for provided session_id")
    }
    broker := engine.broker
    if broker == nil {
        return nil, fmt.Errorf("engine exists but broker is nil. HUHH?")
    }
    return broker, nil
}

func NotifyEngineOfAction(session_id string, action_event UserActionEvent) error {
    engine, ok := engine_manager[session_id]
    if !ok {
        return fmt.Errorf("engine does not exist for provided session_id at this point in time")
    }

    engine.userActionsChan <- action_event

    return nil
}


type MonopolyEngine struct {
    session_id string
    broker *handlers.SseBroker
    userActionsChan chan UserActionEvent // using SseEvent as a template for now
    userActionsChanMu sync.Mutex
}

func SetupNewMonopolyEngine(session_id string) {
    log := log.Logger.With().Str("session_id", session_id).Logger()

    engine := MonopolyEngine {
        session_id: session_id,
        broker: handlers.NewSseBroker(),
        userActionsChan: make(chan UserActionEvent),
    }

    engine_manager_mu.Lock()
    engine_manager[session_id] = &engine
    engine_manager_mu.Unlock()

    ctx := context.Background()

    // infinite loop so we can recover from errors
    for {

        db, err := internaldb.CreateDbPoolConnection(ctx, log)
        if err != nil {
            log.Error().Err(err).Msg("failed to connect to database")
            return
        }

        runMonopolyEngine(log, &engine, db)

    }
}

func runMonopolyEngine(log zerolog.Logger, e *MonopolyEngine, db *pgxpool.Pool) {
    log.Info().Msgf("started monopoly engine for session id: %v", e.session_id)

    for {

        action := <-e.userActionsChan
        log.Trace().Msgf("got action event: %v with data: %v", action.Event, action.Data)

    }

}




