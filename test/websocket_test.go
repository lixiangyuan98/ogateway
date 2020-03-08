package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/lixiangyuan98/ogateway/model"
)

func TestWebsocket(t *testing.T) {
	u := url.URL{

		Scheme: "ws",
		Host:   "127.0.0.1:80",
		Path:   "/ws",
	}
	dialer := &websocket.Dialer{}
	header := http.Header{}
	header.Set("username", "lee")
	header.Set("password", "123456")
	conn, _, err := dialer.Dial(u.String(), header)
	if err != nil {
		t.Fatalf("%v\n", err)
		return
	}
	defer conn.Close()

	go func() {
		for {
			_, m, err := conn.ReadMessage()
			if err != nil {
				log.Printf("read error: %v\n", err)
				return
			}
			log.Printf("recv %v\n", string(m))
		}
	}()

	timer := time.NewTimer(time.Second * 50)
	interval := time.NewTimer(time.Second * 10)

	for {
		select {
		case <-timer.C:
			t.Logf("no error after 3 minutes\n")
			return
		case <-interval.C:
			m, _ := json.Marshal(model.VideoSendRequest{
				Method: "Send",
				Host:   "127.0.0.1",
				Src:    "1.mov",
				Dest:   "127.0.0.1",
				Port:   9999,
			})
			err := conn.WriteMessage(websocket.TextMessage, m)
			if err != nil {
				t.Fatalf("%v\n", err)
				return
			}
			interval.Reset(10 * time.Second)
		}
	}
}
