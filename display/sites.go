package display

import "github.com/c0deaddict/neon-display/display/ws_proto"

type Site struct {
	SiteTitle string `json:"title"`
	SiteOrder int    `json:"order"`
	Url       string `json:"url"`
}

func (s Site) Title() string {
	return s.SiteTitle
}

func (s Site) Order() int {
	return s.SiteOrder
}

func (s Site) Show(t contentTarget) error {
	cmd := ws_proto.OpenSite{
		Title: s.SiteTitle,
		Url:   s.Url,
	}

	msg, err := ws_proto.MakeCommandMessage(ws_proto.OpenSiteCommand, cmd)
	if err != nil {
		return err
	}

	return t.sendMessage(*msg)
}
