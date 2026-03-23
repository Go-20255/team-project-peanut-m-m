package handlers

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "sync"

    "github.com/rs/zerolog"
)

type SseClient struct {
    ID      string
    MsgChan chan []byte
}

type SseBroker struct {
    clients map[*SseClient]struct{}
    mu      sync.RWMutex
}

func NewSseBroker() *SseBroker {
    return &SseBroker{
        clients: make(map[*SseClient]struct{}),
    }
}

func (b *SseBroker) AddClient(client *SseClient) {
    b.mu.Lock()
    b.clients[client] = struct{}{}
    b.mu.Unlock()
}

func (b *SseBroker) RemoveClient(client *SseClient) {
    b.mu.Lock()
    delete(b.clients, client)
    close(client.MsgChan)
    b.mu.Unlock()
}

func (b *SseBroker) Broadcast(log zerolog.Logger, msgObj any) {
    msg, err := json.Marshal(msgObj)
    if err != nil {
        log.Error().Err(err).Msg("failed to marshal object for sse")
        return
    }

    b.mu.RLock()
    defer b.mu.RUnlock()

    for client := range b.clients {
        select {
        case client.MsgChan <- msg:
        default:
            log.Warn().
                Str("client_id", client.ID).
                Msg("dropping sse message for slow client")
        }
    }
}

type SseEvent struct {
    ID      []byte
    Data    []byte
    Event   []byte
    Retry   []byte
    Comment []byte
}

func (ev *SseEvent) MarshalTo(w io.Writer) error {
    if len(ev.Data) == 0 && len(ev.Comment) == 0 {
        return nil
    }

    if len(ev.Data) > 0 {
        if len(ev.ID) > 0 {
            if _, err := fmt.Fprintf(w, "id: %s\n", ev.ID); err != nil {
                return err
            }
        }

        sd := bytes.Split(ev.Data, []byte("\n"))
        for _, line := range sd {
            if _, err := fmt.Fprintf(w, "data: %s\n", line); err != nil {
                return err
            }
        }

        if len(ev.Event) > 0 {
            if _, err := fmt.Fprintf(w, "event: %s\n", ev.Event); err != nil {
                return err
            }
        }

        if len(ev.Retry) > 0 {
            if _, err := fmt.Fprintf(w, "retry: %s\n", ev.Retry); err != nil {
                return err
            }
        }
    }

    if len(ev.Comment) > 0 {
        if _, err := fmt.Fprintf(w, ": %s\n", ev.Comment); err != nil {
            return err
        }
    }

    _, err := fmt.Fprint(w, "\n")
    return err
}

func PrepareSseHeaders(w http.ResponseWriter) {
    h := w.Header()
    h.Set("Content-Type", "text/event-stream")
    h.Set("Cache-Control", "no-cache")
    h.Set("Connection", "keep-alive")
    h.Set("X-Accel-Buffering", "no")
}

func WriteSseEvent(
    w http.ResponseWriter,
    eventName string,
    data []byte,
) error {
    ev := SseEvent{
        Event: []byte(eventName),
        Data:  data,
    }
    if err := ev.MarshalTo(w); err != nil {
        return err
    }

    flusher, ok := w.(http.Flusher)
    if !ok {
        return fmt.Errorf("response writer does not support flushing")
    }
    flusher.Flush()

    return nil
}

func WriteSseComment(w http.ResponseWriter, comment string) error {
    ev := SseEvent{
        Comment: []byte(comment),
    }
    if err := ev.MarshalTo(w); err != nil {
        return err
    }

    flusher, ok := w.(http.Flusher)
    if !ok {
        return fmt.Errorf("response writer does not support flushing")
    }
    flusher.Flush()

    return nil
}

