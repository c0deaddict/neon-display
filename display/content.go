package display

import (
	"fmt"
	"sort"

	"github.com/c0deaddict/neon-display/display/ws_proto"
	"github.com/rs/zerolog/log"
)

type contentTarget interface {
	sendMessage(msg ws_proto.ServerMessage) error
}

type content interface {
	Title() string
	Order() int
	Show(t contentTarget) error
}

type contentList []content

// implement sort.Interface
func (a contentList) Len() int      { return len(a) }
func (a contentList) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a contentList) Less(i, j int) bool {
	if a[i].Order() == a[j].Order() {
		return a[i].Title() < a[j].Title()
	} else {
		return a[i].Order() < a[j].Order()
	}
}

func (list contentList) Find(title string) (int, bool) {
	for i, content := range list {
		if content.Title() == title {
			return i, true
		}
	}
	return 0, false
}

func (d *Display) listContent() contentList {
	list := make([]content, 0)

	albums, err := d.readAlbums()
	if err != nil {
		log.Error().Err(err).Msg("read albums")
	} else {
		for _, album := range albums {
			list = append(list, album)
		}
	}

	for _, site := range d.config.Sites {
		list = append(list, &site)
	}

	result := contentList(list)
	sort.Sort(result)
	return result
}

func (d *Display) initContent() error {
	list := d.listContent()
	if list.Len() == 0 {
		return fmt.Errorf("no content found")
	}

	if index, ok := list.Find(d.config.InitTitle); ok {
		d.currentContent = list[index]
		log.Info().Msgf("starting with content: %s", d.currentContent.Title())
	} else {
		d.currentContent = list[0]
		log.Warn().Msgf("init content not found: %s", d.config.InitTitle)
	}

	return nil
}

func (d *Display) setContent(content content) {
	d.currentContent = content
	err := content.Show(d)
	if err != nil {
		log.Error().Err(err).Msg("show content")
	}
}

func (d *Display) contentStep(step int) {
	list := d.listContent()
	if list.Len() == 0 {
		log.Error().Msg("no content found")
		return
	}

	var c content
	if index, ok := list.Find(d.currentContent.Title()); ok {
		c = list[(index+step)%list.Len()]
	} else {
		c = list[0]
	}

	d.setContent(c)
}

func (d *Display) prevContent() {
	d.contentStep(-1)
}

func (d *Display) nextContent() {
	d.contentStep(1)
}
