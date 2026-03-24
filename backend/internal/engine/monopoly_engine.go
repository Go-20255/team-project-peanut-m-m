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
	engineManager   = map[string]*MonopolyEngine{}
	engineManagerMu sync.Mutex
)

type UserActionEvent struct {
	Event string
	Data  string
}

func GetEngineBroker(sessionId string) (*handlers.SseBroker, error) {
	engine := engineManager[sessionId]
	if engine == nil {
		return nil, fmt.Errorf("engine is nil for provided session_id")
	}
	broker := engine.broker
	if broker == nil {
		return nil, fmt.Errorf("engine exists but broker is nil. HUHH?")
	}
	return broker, nil
}

func NotifyEngineOfAction(sessionId string, actionEvent UserActionEvent) error {
	engine, ok := engineManager[sessionId]
	if !ok {
		return fmt.Errorf("engine does not exist for provided session_id at this point in time")
	}

	engine.userActionsChan <- actionEvent

	return nil
}

type MonopolyEngine struct {
	sessionId         string
	broker            *handlers.SseBroker
	userActionsChan   chan UserActionEvent // using SseEvent as a template for now
	userActionsChanMu sync.Mutex
}

func SetupNewMonopolyEngine(sessionId string) {
	log := log.Logger.With().Str("session_id", sessionId).Logger()

	engine := MonopolyEngine{
		sessionId:       sessionId,
		broker:          handlers.NewSseBroker(),
		userActionsChan: make(chan UserActionEvent),
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

		runMonopolyEngine(log, &engine, db)

	}
}

func runMonopolyEngine(log zerolog.Logger, e *MonopolyEngine, db *pgxpool.Pool) {
	log.Info().Msgf("started monopoly engine for session id: %v", e.sessionId)

	for {

		action := <-e.userActionsChan
		log.Trace().Msgf("got action event: %v with data: %v", action.Event, action.Data)

	}

}
