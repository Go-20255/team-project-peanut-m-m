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

func getExtraRollPlayerId(e *internal.MonopolyEngine) *int {
	for playerId, allowed := range e.ExtraRollAllowed {
		if allowed {
			id := playerId
			return &id
		}
	}

	return nil
}

// EmitInitialGameBoardData gathers the full game board state and broadcasts it
// to connected clients.
// This includes the current turn, all players with their owned properties,
// and all tiles on the board. Use this when a client first joins and needs
// the complete game state.
func EmitInitialGameBoardData(log zerolog.Logger, ctx context.Context, e *internal.MonopolyEngine, tx *pgxpool.Tx, data struct {
    Id         int
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
    board_data.ExtraRollPlayerId = getExtraRollPlayerId(e)
    board_data.PendingCardDraw = e.PendingCardDraw
    board_data.DrawnCard = e.PendingDrawnCard
    board_data.PendingRent = e.PendingRent
    board_data.PendingPropertyPurchase = e.PendingPropertyPurchase
    board_data.PendingBankPayment = e.PendingBankPayment
    board_data.PendingBankPayout = e.PendingBankPayout
    board_data.PendingExchange = e.PendingExchange
    board_data.PendingTrade = e.PendingTrade

    e.Broker.Broadcast(log, "InitialGameBoardDataEvent", board_data)
    return nil
}

// EmitGameBoardUpdate gathers the latest game state and broadcasts an update
// to connected clients.
// This includes the current turn and player state, but not the full tile list.
// Use this after game actions that change the active board state.
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
    board_update.ExtraRollPlayerId = getExtraRollPlayerId(e)
    board_update.PendingCardDraw = e.PendingCardDraw
    board_update.DrawnCard = e.PendingDrawnCard
    board_update.PendingRent = e.PendingRent
    board_update.PendingPropertyPurchase = e.PendingPropertyPurchase
    board_update.PendingBankPayment = e.PendingBankPayment
    board_update.PendingBankPayout = e.PendingBankPayout
    board_update.PendingExchange = e.PendingExchange
    board_update.PendingTrade = e.PendingTrade

    e.Broker.Broadcast(log, "GameStateUpdateEvent", board_update)
    return nil
}
