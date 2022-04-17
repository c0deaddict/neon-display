package display

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"strconv"

	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"

	"github.com/c0deaddict/neon-display/display/ws_proto"
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
	r.POST("/message", d.showMessage)
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

func (d *Display) showMessage(c *gin.Context) {
	show_msg := ws_proto.ShowMessage{ShowSeconds: 5}
	show_msg.Text = c.Query("text")
	if color, ok := c.GetQuery("color"); ok {
		show_msg.Color = &color
	}
	if str, ok := c.GetQuery("show_seconds"); ok {
		val, err := strconv.Atoi(str)
		if err == nil && val > 0 {
			show_msg.ShowSeconds = uint(val)
		}
	}

	msg, err := ws_proto.MakeCommandMessage(ws_proto.ShowMessageCommand, show_msg)
	if err != nil {
		log.Error().Err(err).Msg("make show message command")
	}

	d.broadcast(*msg)
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
		if websocket.IsCloseError(err, websocket.CloseGoingAway,
			websocket.CloseNormalClosure,
			websocket.CloseNoStatusReceived) {
			log.Info().Msgf("client closed connection %p", conn)
			break
		}
		if err != nil {
			log.Error().Err(err).Msg("read ws message")
			break
		}
		if messageType == websocket.TextMessage {
			d.handleMessage(conn, message)
		} else {
			log.Warn().Msgf("ignoring ws message: client %p type %v", conn, messageType)
		}
	}

	d.removeClient(conn)
}

func (d *Display) handleMessage(conn *websocket.Conn, message []byte) {
	log.Info().Msgf("client %p send message %s", conn, string(message))

	var req ws_proto.Request
	err := json.Unmarshal(message, &req)
	if err != nil {
		log.Error().Err(err).Msgf("message from client %p", conn)
		return
	}

	resp := ws_proto.Response{
		RequestId: req.Id,
		Ok:        true,
		Error:     nil,
	}
	msg, err := ws_proto.MakeServerMessage(ws_proto.ResponseMessage, resp)
	if err != nil {
		log.Error().Err(err).Msg("json marshal response message")
		return
	}

	d.mu.Lock()
	defer d.mu.Unlock()
	err = conn.WriteJSON(msg)
	if err != nil {
		log.Warn().Err(err).Msgf("send message to client %p", conn)
	}
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

func (d *Display) broadcast(msg ws_proto.ServerMessage) error {
	log.Info().Msgf("broadcasting %v", msg)

	d.mu.Lock() // also ensures there is no more than one websocket writer.
	defer d.mu.Unlock()

	for _, c := range d.clients {
		err := c.WriteJSON(msg)
		if err != nil {
			log.Warn().Err(err).Msgf("send message to client %p", c)
		}
	}

	return nil
}
