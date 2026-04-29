package game_state_handlers

import (
    "fmt"
    "monopoly-backend/handlers"
    "monopoly-backend/internal"
    monopolyengine "monopoly-backend/internal/engine"
    "monopoly-backend/util"
    "net/http"
    "strconv"
    "time"

    "github.com/labstack/echo/v4"
)

func JoinLiveGameHandler(c echo.Context) error {
    claims, err := util.GetPlayerJwtClaims(c)
    if err != nil {
        return c.String(http.StatusUnauthorized, err.Error())
    }

    sessionId := claims.SessionId
    playerId := strconv.Itoa(claims.PlayerId)
    playerName := claims.PlayerName

    w := c.Response().Writer
    handlers.PrepareSseHeaders(w)

    client := &handlers.SseClient{
        ID:      fmt.Sprintf("%v-%v", playerName, playerId),
        MsgChan: make(chan handlers.SseBroadcastMessage, 32),
        CommentChan: make(chan handlers.SseCommentMessage, 32),
    }

    broker, err := monopolyengine.GetEngineBroker(sessionId)
    if err != nil {
        return c.String(http.StatusBadRequest, "no monopoly engine running for session_id")
    }

    broker.AddClient(client)
    defer broker.RemoveClient(client)

    // notify engine to run the connection event action and wait for the result
    res, err := monopolyengine.NotifyEngineOfAction(sessionId, internal.UserActionEvent{
        Event: "ConnectionEvent",
        Data: struct {
            Id int
            PlayerName string
            SessionId string
        }{
            Id: claims.PlayerId,
            PlayerName: playerName,
            SessionId: sessionId,
        },
        ReturnChan: make(chan internal.UserActionStatus),
    })
    if err != nil {
        return c.String(http.StatusInternalServerError, err.Error())
    }

    if res.Status != http.StatusOK {
        return c.String(res.Status, res.Msg)
    }

    if err := handlers.WriteSseComment(w, "connected"); err != nil {
        return err
    }

    // keep alive heart beat for connection
    heartbeat := time.NewTicker(25 * time.Second)
    defer heartbeat.Stop()

    ctx := c.Request().Context()

    for {
        select {
        case <-ctx.Done():
            // notify engine to run the disconnect event action and wait for the result
            res, err := monopolyengine.NotifyEngineOfAction(sessionId, internal.UserActionEvent{
                Event: "DisconnectEvent",
                Data: struct {
                    Id int
                    PlayerName string
                    SessionId string
                }{
                    Id: claims.PlayerId,
                    PlayerName: playerName,
                    SessionId: sessionId,
                },
                ReturnChan: make(chan internal.UserActionStatus),
            })
            if err != nil {
                return c.String(http.StatusInternalServerError, err.Error())
            }

            if res.Status != http.StatusOK {
                return c.String(res.Status, res.Msg)
            }
            return nil

        case msg, ok := <-client.MsgChan: // monitor for messages we need to broadcast to user
            if !ok {
                return nil
            }

            if err := handlers.WriteSseEvent(w, msg.EventName, msg.MsgObj); err != nil {
                return err
            }

        case msg, ok := <-client.CommentChan: // monitor for comments we need to broadcast to user
            if !ok {
                return nil
            }
            if err := handlers.WriteSseComment(w, msg.Comment); err != nil {
                return err
            }

        case <-heartbeat.C:
            if err := handlers.WriteSseComment(w, "keepalive"); err != nil {
                return nil
            }
        }
    }

}
