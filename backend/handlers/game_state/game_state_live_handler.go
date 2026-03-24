package game_state_handlers

import (
	"fmt"
	"monopoly-backend/handlers"
	monopolyengine "monopoly-backend/internal/engine"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

func JoinLiveGameHandler(c echo.Context) error {

	sessionId := c.QueryParam("session_id")

	playerId := c.FormValue("player_id")
	playerName := c.FormValue("player_name")

	w := c.Response().Writer
	handlers.PrepareSseHeaders(w)

	client := &handlers.SseClient{
		ID:      fmt.Sprintf("%v-%v", playerName, playerId),
		MsgChan: make(chan handlers.SseBroadcastMessage, 32),
	}

	broker, err := monopolyengine.GetEngineBroker(sessionId)
	if err != nil {
		return c.String(http.StatusBadRequest, "no monopoly engine running for session_id")
	}

	broker.AddClient(client)
	defer broker.RemoveClient(client)

	if err := handlers.WriteSseComment(w, "connected"); err != nil {
		return err
	}

	// keep alive heart beat for connection
	heartbeat := time.NewTicker(25 * time.Second)
	defer heartbeat.Stop()

	ctx := c.Request().Context()

	if err := monopolyengine.NotifyEngineOfAction(sessionId, monopolyengine.UserActionEvent{
		Event: "ConnectionEvent",
		Data:  fmt.Sprintf("%v joined game", playerName),
	}); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case msg, ok := <-client.MsgChan: // monitor for messages we need to broadcast to user
			if !ok {
				return nil
			}

			if err := handlers.WriteSseEvent(w, msg.EventName, msg.MsgObj); err != nil {
				return err
			}
		case <-heartbeat.C:
			if err := handlers.WriteSseComment(w, "keepalive"); err != nil {
				return nil
			}
		}
	}

}
