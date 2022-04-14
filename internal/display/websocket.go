package display

import (
	"io/fs"
	"log"

	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/websocket"

	"github.com/c0deaddict/neon-display/frontend"
)

func StartWebsocket() {
	mux := http.NewServeMux()
	fsys, err := fs.Sub(frontend.Assets, "dist")
	if err != nil {
		log.Fatalln(err)
	}
	fileServer := http.FileServer(http.FS(fsys))
	// fileServer := http.FileServer(http.Dir("frontend/dist"))
	mux.Handle("/", fileServer)
	// TODO: file server for photos on /photo ?
	// Handle all other requests
	mux.HandleFunc("/ws", websocketHandler)
	err = http.ListenAndServe(":4000", handlers.LoggingHandler(os.Stdout, mux))
	log.Fatalln(err)
}

func websocketHandler(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{}
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error during connection upgrade:", err)
		return
	}
	defer conn.Close()

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			break
		} else {
			log.Println(messageType, message)
		}

		err = conn.WriteMessage(websocket.TextMessage, []byte("hello"))
		if err != nil {
			log.Println(err)
			break
		}
	}
}
