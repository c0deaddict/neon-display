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

func (s Site) Show() (*ws_proto.ShowContent, error) {
	site := ws_proto.SiteContent{
		Url: s.Url,
	}

	return ws_proto.MakeShowContentMessage(ws_proto.SiteContentType, s.SiteTitle, site)
}
