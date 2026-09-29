package websocket

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	gorillawebsocket "github.com/gorilla/websocket"
)

func TestConnectionConcurrentWrites(t *testing.T) {
	const (
		writerCount     = 16
		writesPerWriter = 16
		payloadSize     = 64 * 1024
	)

	serverResult := make(chan error, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := (&gorillawebsocket.Upgrader{
			CheckOrigin: func(*http.Request) bool { return true },
		}).Upgrade(w, r, nil)
		if err != nil {
			serverResult <- err
			return
		}
		defer conn.Close()
		if err := conn.SetReadDeadline(time.Now().Add(10 * time.Second)); err != nil {
			serverResult <- err
			return
		}

		for i := 0; i < writerCount*writesPerWriter; i++ {
			messageType, payload, err := conn.ReadMessage()
			if err != nil {
				serverResult <- err
				return
			}
			if messageType != gorillawebsocket.BinaryMessage {
				serverResult <- fmt.Errorf("unexpected WebSocket message type: %d", messageType)
				return
			}
			if len(payload) != payloadSize {
				serverResult <- fmt.Errorf("unexpected payload size: got %d, want %d", len(payload), payloadSize)
				return
			}
		}
		serverResult <- nil
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	wsConn, _, err := gorillawebsocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	conn := newConnection(wsConn, wsConn.RemoteAddr())
	defer conn.Close()

	start := make(chan struct{})
	writeErrors := make(chan error, writerCount*writesPerWriter)
	var writers sync.WaitGroup
	for writerID := 0; writerID < writerCount; writerID++ {
		writers.Add(1)
		go func(writerID int) {
			defer writers.Done()
			payload := bytes.Repeat([]byte{byte(writerID)}, payloadSize)
			<-start
			for i := 0; i < writesPerWriter; i++ {
				if _, err := conn.Write(payload); err != nil {
					writeErrors <- err
					return
				}
			}
		}(writerID)
	}

	close(start)
	writers.Wait()
	close(writeErrors)
	for err := range writeErrors {
		t.Errorf("concurrent write failed: %v", err)
	}
	if err := <-serverResult; err != nil {
		t.Fatalf("server failed while reading concurrent writes: %v", err)
	}
}
