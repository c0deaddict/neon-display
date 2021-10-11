package main

import (
	"io/fs"
	"log"

	"flag"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/websocket"
	"github.com/stianeikeland/go-rpio/v4"

	"github.com/c0deaddict/neon-display/frontend"
	"github.com/peterbourgon/ff"
)

var (
	portName      = flag.String("port", "/dev/ttyACM0", "Serial port to read from")
	natsSubject   = flag.String("nats-subject", "sensors.power.emon.radon", "Nats subject to publish on")
	queueCapacity = flag.Uint("queue", 3600, "Queue capacity between serial reader and nats writer")
)

// https://github.com/jgarff/rpi_ws281x
// https://github.com/rpi-ws281x/rpi-ws281x-go

func main() {
	ff.Parse(flag.CommandLine, os.Args[1:],
		ff.WithEnvVarPrefix("NEON"),
		ff.WithConfigFileFlag("config"),
		ff.WithConfigFileParser(ff.PlainParser))

	// go metrics.Run()

	mux := http.NewServeMux()
	fsys, err := fs.Sub(frontend.Assets, "dist")
	if err != nil {
		log.Fatalln(err)
	}
	fileServer := http.FileServer(http.FS(fsys))
	// fileServer := http.FileServer(http.Dir("frontend/dist"))
	// TODO: add websocket for comm.
	// TODO: add logging for http requests
	mux.Handle("/", fileServer)
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

func test() {
	// TODO: look at https://github.com/stianeikeland/go-rpio#using-without-root
	err := rpio.Open()
	if err != nil {
		log.Fatalln(err)
	}
	defer rpio.Close()

	pin := rpio.Pin(10)
	pin.PullUp()
	pin.Input()
}
