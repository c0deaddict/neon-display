package display

import (
	"fmt"
	"io/fs"

	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"

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
		log.Fatal().Err(err).Msg("loading frontend files")
	}

	r := gin.Default()

	// TODO: could use https://github.com/gin-contrib/logger

	r.StaticFS("/photo", http.Dir(d.config.PhotosPath))
	r.GET("/ws", func(c *gin.Context) {
		d.websocketHandler(c.Writer, c.Request)
	})
	r.GET("/metrics", prometheusHandler())
	r.GET("/event/:name", d.userEvent)
	r.NoRoute(func(c *gin.Context) {
		c.FileFromFS(c.Request.URL.Path, http.FS(fsys))
	})

	listen := fmt.Sprintf("%s:%d", d.config.WebBind, d.config.WebPort)
	err = r.Run(listen)
	if err != nil {
		log.Fatal().Err(err).Msg("web server start")
	}

}

func (d *Display) userEvent(c *gin.Context) {
	name := c.Param("name")
	log.Printf("event %s", name)
	c.Status(http.StatusNoContent)
}

func (d *Display) websocketHandler(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{}
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().Err(err).Msg("connection upgrade")
		return
	}
	defer conn.Close()
	d.addClient(conn)

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Warn().Err(err).Msg("read ws message")
			break
		} else {
			log.Info().Msgf("received ws message: client %p type %v data %v", conn, messageType, message)
		}

		err = conn.WriteMessage(websocket.TextMessage, []byte("hello"))
		if err != nil {
			log.Warn().Err(err).Msg("send ws message")
			break
		}
	}

	d.removeClient(conn)
}

func (d *Display) addClient(conn *websocket.Conn) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.clients = append(d.clients, conn)
	log.Info().Msgf("adding client %p", conn)
}

func (d *Display) removeClient(conn *websocket.Conn) {
	d.mu.Lock()
	defer d.mu.Unlock()
	for i, other := range d.clients {
		if other == conn {
			d.clients = append(d.clients[:i], d.clients[i+1:]...)
			log.Info().Msgf("removed client %p", conn)
			return
		}
	}

	log.Warn().Msgf("remove client: %p is not found", conn)
}
