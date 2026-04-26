package events

import (
	"context"
	"monopoly-backend/internal"
	internaldb_game_state "monopoly-backend/internal/db/game_state"
	internaldb_players "monopoly-backend/internal/db/player"
	internaldb_tiles "monopoly-backend/internal/db/tile"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)
func EmitInitialGameBoardData(log zerolog.Logger, ctx context.Context, e *internal.MonopolyEngine, tx *pgxpool.Tx, data struct {
    Id         string
    PlayerName string
    SessionId  string
}) error {

    // get current turn
    current_turn, err := internaldb_game_state.GetGameStateTurnNumber(
        log,
        ctx,
        tx,
        data.SessionId,
    )
    if err != nil {
        return err
    }

    // get players
    players, err := internaldb_players.GetPlayersInSession(
        log,
        ctx,
        tx,
        data.SessionId,
    )
    if err != nil {
        return err
    }
    // get player infos
    var all_players_info []internal.PlayerInfo
    for _, p := range players {
        player_owned_properties, err := internaldb_players.GetPlayerOwnedProperties(
            log,
            ctx,
            tx,
            p.Id,
            p.SessionId,
        )
        if err != nil {
            return err
        }

        player_info := internal.PlayerInfo{
            Player:          p,
            OwnedProperties: player_owned_properties,
        }
        all_players_info = append(all_players_info, player_info)
    }

    // get tiles
    tiles, err := internaldb_tiles.GetAllTiles(
        log,
        ctx,
        tx,
        data.SessionId,
    )
    if err != nil {
        return err
    }

    var board_data internal.GameBoardData
    board_data.CurrentTurn = current_turn
    board_data.Players = all_players_info
    board_data.Tiles = tiles

    e.Broker.Broadcast(log, "InitialGameBoardDataEvent", board_data)
    return nil
}


func EmitGameBoardUpdate(log zerolog.Logger, ctx context.Context, e *internal.MonopolyEngine, tx *pgxpool.Tx) error {
    // get current turn
    current_turn, err := internaldb_game_state.GetGameStateTurnNumber(
        log,
        ctx,
        tx,
        e.SessionId,
    )
    if err != nil {
        return err
    }

    // get players
    players, err := internaldb_players.GetPlayersInSession(
        log,
        ctx,
        tx,
        e.SessionId,
    )
    if err != nil {
        return err
    }
    // get player infos
    var all_players_info []internal.PlayerInfo
    for _, p := range players {
        player_owned_properties, err := internaldb_players.GetPlayerOwnedProperties(
            log,
            ctx,
            tx,
            p.Id,
            p.SessionId,
        )
        if err != nil {
            return err
        }

        player_info := internal.PlayerInfo{
            Player:          p,
            OwnedProperties: player_owned_properties,
        }
        all_players_info = append(all_players_info, player_info)
    }

    var board_update internal.GameStateUpdate
    board_update.CurrentTurn = current_turn
    board_update.Players = all_players_info

    e.Broker.Broadcast(log, "GameStateUpdateEvent", board_update)
    return nil
}


