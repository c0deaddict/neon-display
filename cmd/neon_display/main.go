package main

import (
	"io/fs"
	"log"

	"flag"
	"net/http"
	"os"

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
	mux.Handle("/", fileServer)
	// Handle all other requests
	mux.HandleFunc("/ws", func(w http.ResponseWriter, req *http.Request) {
		var path = req.URL.Path
		log.Println("Serving request for path", path)
		w.Header().Add("Content-Type", "text/plain")
		w.Write([]byte("hello"))
	})
	err = http.ListenAndServe(":4000", mux)
	log.Fatalln(err)
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
