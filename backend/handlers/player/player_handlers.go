package players_handlers

import (
	internaldb_game_state "monopoly-backend/internal/db/game_state"
	internaldb_players "monopoly-backend/internal/db/players"
	"monopoly-backend/util"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)


func CreatePlayerHandler(c echo.Context) error {
    log := util.GetRequestLogger(c)
    ctx := c.Request().Context()

    name := c.FormValue("player_name")
    if name == "" {
        return c.String(http.StatusBadRequest, "missing player name")
    }

    session_id_str := c.FormValue("session_id")
    session_id, err := strconv.Atoi(session_id_str)
    if err != nil {
        return c.String(http.StatusBadRequest, "failed to parse invalid session_id")
    }

    tx := c.Get("tx").(*pgxpool.Tx)
    err = internaldb_game_state.GameStateExists(log, ctx, tx, session_id) 
    if err != nil {
        return c.String(http.StatusBadRequest, "session_id does not exist")
    }

    err = internaldb_players.CreatePlayerDB(log, ctx, tx, name, session_id)
    if err != nil {
        return c.String(http.StatusInternalServerError, "failed to create player in db")
    }

    return c.String(http.StatusOK, "created player")
}
