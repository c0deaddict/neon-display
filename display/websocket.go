package display

import (
	"io/fs"
	"log"

	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/c0deaddict/neon-display/frontend"
)

func prometheusHandler() gin.HandlerFunc {
	h := promhttp.Handler()

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

func (d *Display) StartWebsocket() {
	fsys, err := fs.Sub(frontend.Assets, "dist")
	if err != nil {
		log.Fatalln(err)
	}

	r := gin.Default()

	r.StaticFS("/photo", http.Dir("./photos")) // TODO: configurable photo dir
	r.GET("/ws", func(c *gin.Context) {
		websocketHandler(c.Writer, c.Request)
	})
	r.GET("/metrics", prometheusHandler())
	r.NoRoute(func(c *gin.Context) {
		c.FileFromFS(c.Request.URL.Path, http.FS(fsys))
	})

	err = r.Run() // listen and serve on 0.0.0.0:8080
	if err != nil {
		log.Fatalln(err)
	}

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
