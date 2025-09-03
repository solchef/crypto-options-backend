package config

import (
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // allow all for testing
}

// func ServeWS(w http.ResponseWriter, r *http.Request) {
// 	conn, err := upgrader.Upgrade(w, r, nil)
// 	if err != nil {
// 		return
// 	}
// 	WSHub.AddClient(conn)

// 	// Clean up when disconnected
// 	defer WSHub.RemoveClient(conn)

// 	for {
// 		// Just keep connection alive, ignore input
// 		if _, _, err := conn.ReadMessage(); err != nil {
// 			break
// 		}
// 	}
// }

func ServeWS(w http.ResponseWriter, r *http.Request) {
	// TODO: extract userID from JWT in header/cookie
	userID := uint(5) // placeholder for testing

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	WSHub.AddClient(userID, conn)
	defer WSHub.RemoveClient(userID, conn)

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}
