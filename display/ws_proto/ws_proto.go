package ws_proto

import (
	"encoding/json"
)

type RequestMethod string
type ServerMessageType string
type CommandType string

const (
	PingMethod RequestMethod = "ping"

	ResponseMessage ServerMessageType = "response"
	CommandMessage  ServerMessageType = "command"

	StartSlideshowCommand CommandType = "start_slideshow"
	StopSlideshowCommand  CommandType = "stop_slideshow"
	OpenUrlCommand        CommandType = "open_url"
	ShowMessageCommand    CommandType = "show_message"
	ReloadCommand         CommandType = "reload"
)

type Request struct {
	Id     string          `json:"id"`
	Method RequestMethod   `json:"method"`
	Params json.RawMessage `json:"params"`
}

type ServerMessage struct {
	Type ServerMessageType `json:"type"`
	Data json.RawMessage   `json:"data"`
}

type Response struct {
	RequestId string  `json:"request_id"`
	Ok        bool    `json:"ok"`
	Error     *string `json:"error,omitempty"`
}

type Command struct {
	Type CommandType     `json:"type"`
	Data json.RawMessage `json:"data"`
}

type Photo struct {
	ImageUrl string `json:"url"`
	Caption  string `json:"caption"`
	Date     string `json:"date"`
}

type StartSlideshow struct {
	AlbumTitle      string  `json:"album_title"`
	DelaySeconds    uint    `json:"delay_seconds"`
	BackgroundColor *string `json:"background_color,omitempty"`
	Photos          []Photo `json:"photos"`
}

type OpenUrl struct {
	Url string `json:"url"`
}

type ShowMessage struct {
	Text        string  `json:"text"`
	Color       *string `json:"color,omitempty"`
	ShowSeconds uint    `json:"show_seconds"`
}

func MakeServerMessage(typ ServerMessageType, data interface{}) (*ServerMessage, error) {
	raw_data, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return &ServerMessage{Type: typ, Data: json.RawMessage(raw_data)}, nil
}

func MakeCommandMessage(typ CommandType, data interface{}) (*ServerMessage, error) {
	raw_data, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	command := Command{Type: typ, Data: json.RawMessage(raw_data)}
	return MakeServerMessage(CommandMessage, command)
}
