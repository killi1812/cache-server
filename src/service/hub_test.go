package service

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

var upgrader = websocket.Upgrader{}

func TestHub(t *testing.T) {
	hub := NewHub()
	
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, _ := upgrader.Upgrade(w, r, nil)
		hub.Register("test-agent", conn)
	}))
	defer s.Close()

	t.Run("Register and Notify", func(t *testing.T) {
		u := "ws" + strings.TrimPrefix(s.URL, "http")
		client, _, err := websocket.DefaultDialer.Dial(u, nil)
		assert.NoError(t, err)
		defer client.Close()

		// hub.Register happens in server handler
		
		msg := map[string]string{"test": "data"}
		err = hub.NotifyAgent("test-agent", msg)
		assert.NoError(t, err)

		var received map[string]string
		err = client.ReadJSON(&received)
		assert.NoError(t, err)
		assert.Equal(t, "data", received["test"])
	})

	t.Run("Unregister", func(t *testing.T) {
		hub.Unregister("test-agent")
		err := hub.NotifyAgent("test-agent", "fail")
		assert.Error(t, err)
	})
}
