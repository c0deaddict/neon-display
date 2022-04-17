package display

import "github.com/c0deaddict/neon-display/display/ws_proto"

type Site struct {
	Title string `json:"title"`
	Url   string `json:"url"`
}

func (s Site) GetTitle() string {
	return s.Title
}

func (s Site) Show(t contentTarget) error {
	cmd := ws_proto.OpenSite{
		Title: s.Title,
		Url:   s.Url,
	}

	msg, err := ws_proto.MakeCommandMessage(ws_proto.OpenSiteCommand, cmd)
	if err != nil {
		return err
	}

	return t.sendMessage(*msg)
}
