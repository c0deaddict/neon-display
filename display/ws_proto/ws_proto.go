package ws_proto

import (
	"encoding/json"
)

type RequestMethod string
type ServerMessageType string
type CommandType string
type ContentType string

const (
	PingMethod RequestMethod = "ping"

	ResponseMessage ServerMessageType = "response"
	CommandMessage  ServerMessageType = "command"

	ShowContentCommand   CommandType = "show_content"
	PauseContentCommand  CommandType = "pause_content"
	ResumeContentCommand CommandType = "resume_content"
	ShowMessageCommand   CommandType = "show_message"
	ReloadCommand        CommandType = "reload"

	PhotosContentType ContentType = "photos"
	SiteContentType   ContentType = "site"
	VideoContentType  ContentType = "video"
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
	ImagePath   string  `json:"image_path"`
	Description *string `json:"description,omitempty"`
	DateTime    *string `json:"datetime,omitempty"`
	Camera      *string `json:"camera,omitempty"`
}

type ShowContent struct {
	Type  ContentType     `json:"type"`
	Title string          `json:"title"`
	Data  json.RawMessage `json:"data"`
}

type PhotosContent struct {
	DelaySeconds    uint    `json:"delay_seconds"`
	BackgroundColor *string `json:"background_color,omitempty"`
	Photos          []Photo `json:"photos"`
}

type SiteContent struct {
	Url string `json:"url"`
}

type VideoContent struct {
	Path string `json:"path"`
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

func MakeShowContentMessage(typ ContentType, title string, data interface{}) (*ShowContent, error) {
	raw_data, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	show := ShowContent{Type: typ, Title: title, Data: json.RawMessage(raw_data)}
	return &show, nil
}
