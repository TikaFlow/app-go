package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
)

func On(event string, action func(ws *websocket.Conn, event string, data any)) error {
	if event == "" {
		return errors.New("event name cannot be empty")
	}
	if action == nil {
		return errors.New("action cannot be nil")
	}

	busLock.Lock()
	defer busLock.Unlock()

	if _, ok := bus[event]; ok {
		bus[event] = action
		return errors.New("event already exists, and will be overwritten")
	}

	bus[event] = action
	return nil
}

func Off(event string) error {
	if event == "" {
		return errors.New("event name cannot be empty")
	}

	busLock.Lock()
	defer busLock.Unlock()
	delete(bus, event)
	return nil
}

func listenWs() {
	http.HandleFunc("/ws", handleWs)
	if webRoot != nil {
		http.Handle("/", http.FileServer(http.FS(webRoot)))
	}

	var ln net.Listener
	var err error
	i := 6239
	for i = 6239; i < 54321; i++ {
		ln, err = net.Listen("tcp", fmt.Sprintf(":%d", i))
		if err == nil {
			break
		}
	}
	_ = os.Setenv("APP_PORT", fmt.Sprintf("%d", i))
	close(appReadyChan)
	fmt.Println("app listening on port", i)
	_ = http.Serve(ln, nil)
}

func handleWs(w http.ResponseWriter, r *http.Request) {
	conn, err := wsUp.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer func(conn *websocket.Conn) {
		_ = conn.Close()
	}(conn)

	slog.Info("new connection", "remote", conn.RemoteAddr())
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			slog.Info("connection closed", "remote", conn.RemoteAddr(), "error", err)
			return
		}

		var message Ping
		if err := json.Unmarshal(msg, &message); err != nil {
			fmt.Println("parse data error:", err)
			continue
		}

		slog.Info("received", "event", message.Event, "data", message.Data)
		busLock.RLock()
		if action, ok := bus[message.Event]; ok {
			action(conn, message.Event, message.Data)
		}
		busLock.RUnlock()
	}
}
