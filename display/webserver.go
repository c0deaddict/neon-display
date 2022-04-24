package display

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"path"
	"strconv"
	"time"

	"net/http"

	"github.com/gin-contrib/logger"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/c0deaddict/neon-display/display/ws_proto"
	"github.com/c0deaddict/neon-display/frontend"
)

type client struct {
	*websocket.Conn
	display *Display
}

func (c client) sendMessage(msg ws_proto.ServerMessage) error {
	c.display.mu.Lock() // ensure there is no more than one websocket writer.
	defer c.display.mu.Unlock()

	err := c.WriteJSON(msg)
	if err != nil {
		log.Warn().Err(err).Msgf("send message to client %p", &c)
	}

	return nil
}

func prometheusHandler() gin.HandlerFunc {
	h := promhttp.Handler()

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

func (d *Display) startWebserver() error {
	fsys, err := fs.Sub(frontend.Assets, "dist")
	if err != nil {
		return fmt.Errorf("loading frontend files: %s", err)
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery()) // 500 on panics

	// Log requests to zerolog.
	r.Use(logger.SetLogger(logger.WithLogger(func(c *gin.Context, out io.Writer, latency time.Duration) zerolog.Logger {
		return log.Logger.
			With().
			Timestamp().
			Int("status", c.Writer.Status()).
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Str("ip", c.ClientIP()).
			Dur("latency", latency).
			Str("user_agent", c.Request.UserAgent()).
			Logger()
	})))

	if d.config.PhotosPath != "" {
		r.StaticFS("/photo", http.Dir(d.config.PhotosPath))
	}

	// Videos must be able to be streamed, Gin's r.StaticFS can't do that.
	// Use http.ServeFile, which does support streaming (ranges).
	// https://stackoverflow.com/questions/63221721/serve-video-with-go-gin
	if d.config.VideosPath != "" {
		r.GET("/video/:name", d.videoHandler)
	}

	r.GET("/ws", func(c *gin.Context) {
		d.websocketHandler(c.Writer, c.Request)
	})
	r.GET("/metrics", prometheusHandler())
	r.GET("/event/:name", d.userEvent)
	r.POST("/message", d.messageHandler)
	r.NoRoute(func(c *gin.Context) {
		c.FileFromFS(c.Request.URL.Path, http.FS(fsys))
	})

	go func() {
		listen := fmt.Sprintf("%s:%d", d.config.WebBind, d.config.WebPort)
		err := r.Run(listen)
		if err != nil {
			log.Fatal().Err(err).Msg("web server start")
		}
	}()

	return nil
}

func (d *Display) userEvent(c *gin.Context) {
	name := c.Param("name")
	log.Printf("event %s", name)
	c.Status(http.StatusNoContent)
}

func (d *Display) messageHandler(c *gin.Context) {
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

	d.showMessage(show_msg)
	c.Status(http.StatusNoContent)
}

func (d *Display) showMessage(msg ws_proto.ShowMessage) {
	cmd, err := ws_proto.MakeCommandMessage(ws_proto.ShowMessageCommand, msg)
	if err != nil {
		log.Error().Err(err).Msg("make show message command")
	}

	d.sendMessage(*cmd)
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

	client := client{conn, d}
	d.addClient(client)

	// Show the current content on the client.
	d.showContentOnTarget(client)

	for {
		messageType, message, err := client.ReadMessage()
		if websocket.IsCloseError(err, websocket.CloseGoingAway,
			websocket.CloseNormalClosure,
			websocket.CloseNoStatusReceived) {
			log.Info().Msgf("client closed connection %p", &client)
			break
		}
		if err != nil {
			log.Error().Err(err).Msg("read ws message")
			break
		}
		if messageType == websocket.TextMessage {
			d.handleMessage(client, message)
		} else {
			log.Warn().Msgf("ignoring ws message: client %p type %v", &client, messageType)
		}
	}

	d.removeClient(client)
}

func (d *Display) handleMessage(c client, message []byte) {
	log.Info().Msgf("client %p send message %s", &c, string(message))

	var req ws_proto.Request
	err := json.Unmarshal(message, &req)
	if err != nil {
		log.Error().Err(err).Msgf("message from client %p", &c)
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
	err = c.WriteJSON(msg)
	if err != nil {
		log.Warn().Err(err).Msgf("send message to client %p", &c)
	}
}

func (d *Display) addClient(c client) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.clients = append(d.clients, c)
	log.Info().Msgf("adding client %p", &c)
}

func (d *Display) removeClient(c client) {
	d.mu.Lock()
	defer d.mu.Unlock()
	for i, other := range d.clients {
		if other == c {
			d.clients = append(d.clients[:i], d.clients[i+1:]...)
			log.Info().Msgf("removed client %p", &c)
			return
		}
	}

	log.Warn().Msgf("remove client: %p is not found", &c)
}

func (d *Display) sendMessage(msg ws_proto.ServerMessage) error {
	d.mu.Lock() // also ensures there is no more than one websocket writer.
	defer d.mu.Unlock()

	for _, c := range d.clients {
		err := c.WriteJSON(msg)
		if err != nil {
			log.Warn().Err(err).Msgf("send message to client %p", &c)
		}
	}

	return nil
}

func (d *Display) videoHandler(c *gin.Context) {
	name := c.Param("name")
	filepath := path.Join(d.config.VideosPath, name)
	http.ServeFile(c.Writer, c.Request, filepath)
}
