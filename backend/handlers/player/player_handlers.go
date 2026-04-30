package players_handlers

import (
    "monopoly-backend/util"
    "net/http"
    "strconv"
	"monopoly-backend/internal"
	internaldbgamestate "monopoly-backend/internal/db/game_state"
	internaldbplayers "monopoly-backend/internal/db/player"
	monopoly_engine "monopoly-backend/internal/engine"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

func CreatePlayerHandler(c echo.Context) error {
    log := util.GetRequestLogger(c)
    ctx := c.Request().Context()

    name := c.FormValue("player_name")
    if name == "" {
        return c.String(http.StatusBadRequest, "missing player_name")
    }

    sessionId := c.FormValue("session_id")
    if sessionId == "" {
        return c.String(http.StatusBadRequest, "missing session_id")
    }

    pieceToken_str := c.FormValue("piece_token")
    if pieceToken_str == "" {
        return c.String(http.StatusBadRequest, "missing piece_token")
    }
    pieceToken, err := strconv.Atoi(pieceToken_str)
    if err != nil {
        return c.String(http.StatusBadRequest, "piece token is not a valid int")
    }

    tx := c.Get("tx").(*pgxpool.Tx)
    exists, err := internaldbgamestate.GameStateExists(log, ctx, tx, sessionId)
    if err != nil {
        return c.String(http.StatusInternalServerError, "failed to query db about game state")
    }

    if !exists {
        return c.String(http.StatusBadRequest, "session_id does not exist")
    }

    //pieceToken, err := util.AssignPlayerToken(log, ctx, tx, sessionId)
    //if err != nil {
        //return c.String(http.StatusInternalServerError, "failed to assign player token")
    //}

    id, err := internaldbplayers.CreatePlayerDB(log, ctx, tx, name, sessionId, pieceToken)
    if err != nil {
        return c.String(http.StatusInternalServerError, "failed to create player in db")
    }

    new_player, err := internaldbplayers.GetPlayer(log, ctx, tx, id, sessionId)
    if err != nil {
        return c.String(http.StatusInternalServerError, "uh oh. something bad happened")
    }

    return c.JSON(http.StatusOK, new_player)
}

func GetPlayersHandler(c echo.Context) error {
    log := util.GetRequestLogger(c)
    ctx := c.Request().Context()

    sessionId := c.QueryParam("session_id")
    if sessionId == "" {
        return c.String(http.StatusBadRequest, "missing session_id")
    }

    tx := c.Get("tx").(*pgxpool.Tx)
    
    players, err := internaldbplayers.GetPlayersInSession(log, ctx, tx, sessionId)
    if err != nil {
        return c.String(http.StatusInternalServerError, "failed to fetch players")
    }

    return c.JSON(http.StatusOK, players)
}

func UpdatePlayerTokenHandler(c echo.Context) error {
    log := util.GetRequestLogger(c)
    ctx := c.Request().Context()

    playerId, _ := strconv.Atoi(c.FormValue("player_id"))
    sessionId := c.FormValue("session_id")
    pieceToken, _ := strconv.Atoi(c.FormValue("piece_token"))

    tx := c.Get("tx").(*pgxpool.Tx)
    err := internaldbplayers.UpdatePlayerPieceToken(log, ctx, tx, playerId, sessionId, pieceToken)
    if err != nil {
        return c.String(http.StatusInternalServerError, "failed to update player token")
    }

    return c.JSON(http.StatusOK, map[string]interface{}{
        "id":           playerId,
        "session_id":   sessionId,
        "piece_token":  pieceToken,
    })
}

func JoinPlayerHandler(c echo.Context) error {
    log := util.GetRequestLogger(c)
    ctx := c.Request().Context()

    playerId_str := c.QueryParam("player_id")
    playerId, err := strconv.Atoi(playerId_str)
    if err != nil {
        return c.String(http.StatusBadRequest, "player id is not an integer")
    }


    name := c.QueryParam("player_name")
    if name == "" {
        return c.String(http.StatusBadRequest, "missing player_name")
    }

    sessionId := c.QueryParam("session_id")
    if sessionId == "" {
        return c.String(http.StatusBadRequest, "missing session_id")
    }

    tx := c.Get("tx").(*pgxpool.Tx)
    exists, err := internaldbgamestate.GameStateExists(log, ctx, tx, sessionId)
    if err != nil {
        return c.String(http.StatusInternalServerError, "failed to query db about game state")
    }

    if !exists {
        return c.String(http.StatusBadRequest, "session_id does not exist")
    }

    playerExists, err := internaldbplayers.CheckPlayerExists(log, ctx, tx, playerId, name, sessionId)
    if err != nil {
        return c.String(http.StatusInternalServerError, "failed to query db about player")
    }

    if !playerExists {
        return c.String(http.StatusBadRequest, "player does not exist")
    }

    jwt, err := util.CreatePlayerJwt(playerId, name, sessionId)
    if err != nil {
        return c.String(http.StatusInternalServerError, "failed to create player auth token")
    }

    c.SetCookie(&http.Cookie{
        Name:    util.PlayerAuthCookieName,
        Value:   jwt,
        Path:    "/",
        Secure:  true,
        Expires: time.Now().Add(24 * time.Hour),
    })

    return c.JSON(http.StatusOK, map[string]interface{}{
        "id":         playerId_str,
        "name":       name,
        "session_id": sessionId,
        "jwt":        jwt,
    })
}



func ReadyUpPlayerHandler(c echo.Context) error {
    claims, err := util.GetPlayerJwtClaims(c)
    if err != nil {
        return c.String(http.StatusUnauthorized, err.Error())
    }

    status_str := c.QueryParam("status")
    status, err := strconv.ParseBool(status_str)
    if err != nil {
        return c.String(http.StatusBadRequest, err.Error())
    }

    res, err := monopoly_engine.NotifyEngineOfAction(claims.SessionId, internal.UserActionEvent{
        Event: "PlayerReadyUpEvent",
        Data: struct {
            SessionId   string
            PlayerId    int
            Status      bool
        }{
            SessionId: claims.SessionId,
            PlayerId: claims.PlayerId,
            Status: status,
        },
        ReturnChan: make(chan internal.UserActionStatus),
    })

    if err != nil {
        return c.String(http.StatusInternalServerError, err.Error())
    }

    if res.Status != http.StatusOK {
        return c.String(res.Status, res.Msg)
    }

    return c.String(http.StatusOK, res.Msg)
}

func EndTurnHandler(c echo.Context) error {
    claims, err := util.GetPlayerJwtClaims(c)
    if err != nil {
        return c.String(http.StatusUnauthorized, err.Error())
    }

    res, err := monopoly_engine.NotifyEngineOfAction(claims.SessionId, internal.UserActionEvent{
        Event: "EndTurnEvent",
        Data: internal.SimpleActionData {
            PlayerId: claims.PlayerId,
            SessionId: claims.SessionId,
        },
        ReturnChan: make(chan internal.UserActionStatus),
    })

    if err != nil {
        return c.String(http.StatusInternalServerError, err.Error())
    }

    if res.Status != http.StatusOK {
        return c.String(res.Status, res.Msg)
    }

    return c.String(http.StatusOK, res.Msg)

}



