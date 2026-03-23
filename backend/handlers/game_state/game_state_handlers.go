package game_state

import (
	internaldbgamestate "monopoly-backend/internal/db/game_state"
	"monopoly-backend/util"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

func JoinGameHandler(c echo.Context) error {
	log := util.GetRequestLogger(c)
	ctx := c.Request().Context()

	code := c.QueryParam("code")
	if code == "" {
		return c.String(http.StatusBadRequest, "missing code")
	}

	tx := c.Get("tx").(*pgxpool.Tx)
	id, err := internaldbgamestate.CheckGameStateCode(log, ctx, tx, code)
	if err != nil {
		return c.String(http.StatusBadRequest, "code does not exist")
	}

	return c.String(http.StatusOK, "joined game: "+id)
}
