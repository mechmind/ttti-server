package session

import (
    "net/http"
    "encoding/json"
    "fmt"
    "log"
    //"net"
)

type Handler struct {
    r *Registry
    shutdown_chan chan bool
}


func (h *Handler) CreateSession(w http.ResponseWriter, r *http.Request) {
    id := r.FormValue("session_id")
    session, err := h.r.CreateSession(id)
    var answer []byte
    if err == nil {
        answer, _ = json.Marshal(map[string]string{"session_id": session.id, "type": "response"})
        log.Printf("http-api: created session '%s'", session.id)
    } else {
        answer, _ = json.Marshal(map[string]string{"type": "error", "message": err.Error()})
    }
    _, _ = w.Write(answer)
}

func (h *Handler) AttachPlayer(w http.ResponseWriter, r *http.Request) {
    session := r.FormValue("session_id")
    if session == "" {
        http.Error(w, `{"type":"error", "message":"session_id is missing"}`, 400)
        return
    }
    player := r.FormValue("player_id")
    if player == "" {
        http.Error(w, `{"type":"error", "message":"player_id is missing"}`, 400)
        return
    }
    player, glyph, err := h.r.AttachPlayer(session, player)
    if err != nil {
        errstr, _ := json.Marshal(map[string]string{"type":"error", "message":err.Error()})
        http.Error(w, string(errstr), 400)
        return
    }
    answer, _ := json.Marshal(map[string]string{"session_id": session, "player_id": player,
    "player_glyph": glyph, "type": "response"})
    log.Printf("http-api: updated session '%s' attached player '%s'", session, player)
    _, _ = w.Write(answer)
}

func (h *Handler) Shutdown(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte{})
    h.shutdown_chan <- true
}

func (h *Handler) Dump(w http.ResponseWriter, r *http.Request) {
    sessions := []Session{}
    for _, session := range h.r.sessions {
        sessions = append(sessions, *session)
    }
    fmt.Fprint(w, sessions)
}

func ServeHttp(addr string, r *Registry) chan bool {
    handler := &Handler{r, make(chan bool)}
    http.HandleFunc("/create_session", handler.CreateSession)
    http.HandleFunc("/attach_player", handler.AttachPlayer)
    http.HandleFunc("/shutdown", handler.Shutdown)
    http.HandleFunc("/dump", handler.Dump)
    go http.ListenAndServe(addr, nil)
    return handler.shutdown_chan
}
