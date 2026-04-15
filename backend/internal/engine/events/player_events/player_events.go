package player_events

import (
	"context"
	"fmt"
	"monopoly-backend/internal"
	internaldb_players "monopoly-backend/internal/db/players"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

func Connected(
    ctx context.Context,
    log zerolog.Logger,
    e *internal.MonopolyEngine,
    action *internal.UserActionEvent,
    tx *pgxpool.Tx,
) (internal.UserActionStatus) {
    action_status := internal.UserActionStatus{
        Status: http.StatusOK,
    }
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
        action_status = internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
        return action_status
    }

    if !player_exists {
        action_status = internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "player does not exist",
        }
        return action_status
    }


    internaldb_players.UpdatePlayerInGameStatus(log, ctx, tx, data.Id, data.SessionId, true)

    // announce to all connected users that another user has joined the game
    e.Broker.Broadcast(log, "ConnectionEvent", fmt.Sprintf("player %v has joined", data.PlayerName))
    log.Trace().Msgf("player %v has joined server", data.PlayerName)
    return action_status
}




