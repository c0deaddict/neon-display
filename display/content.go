package display

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/c0deaddict/neon-display/display/photos"
	"github.com/c0deaddict/neon-display/display/ws_proto"
	"github.com/rs/zerolog/log"
)

type contentTarget interface {
	sendMessage(msg ws_proto.ServerMessage) error
}

type content interface {
	Title() string
	Order() int
	Show() (*ws_proto.ShowContent, error)
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

	albums, err := photos.ReadAlbums(d.config.Photos)
	if err != nil {
		log.Error().Err(err).Msg("read albums")
	} else {
		for _, album := range albums {
			list = append(list, album)
		}
	}

	videos, err := d.readVideos()
	if err != nil {
		log.Error().Err(err).Msg("read videos")
	} else {
		for _, video := range videos {
			list = append(list, video)
		}
	}

	for _, site := range d.config.Sites {
		list = append(list, site)
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

	log.Info().Msg("found content:")
	for _, c := range list {
		log.Info().Msgf("%v \"%s\"", reflect.TypeOf(c), c.Title())
	}

	if d.config.InitTitle != "" {
		if index, ok := list.Find(d.config.InitTitle); ok {
			d.currentContent = list[index]
			log.Info().Msgf("starting with content: %s", d.currentContent.Title())
		} else {
			log.Warn().Msgf("init content not found: %s", d.config.InitTitle)
		}
	}
	if d.currentContent == nil {
		d.currentContent = list[0]
	}

	return nil
}

func (d *Display) showContentOnTarget(t contentTarget) {
	show, err := d.currentContent.Show()
	if err != nil {
		log.Error().Err(err).Msg("show content")
		return
	}

	msg, err := ws_proto.MakeCommandMessage(ws_proto.ShowContentCommand, show)
	if err != nil {
		log.Error().Err(err).Msg("make show command message")
		return
	}

	err = t.sendMessage(*msg)
	if err != nil {
		log.Error().Err(err).Msg("send show content command")
	}
}

func (d *Display) setContent(content content) {
	d.currentContent = content
	d.showContentOnTarget(d)
}

func (d *Display) contentStep(step int) {
	list := d.listContent()
	if list.Len() == 0 {
		log.Error().Msg("no content found")
		return
	}

	var c content
	if index, ok := list.Find(d.currentContent.Title()); ok {
		index = (index + step) % list.Len()
		// Go's modulo can return negative numbers.
		if index < 0 {
			index = list.Len() + index
		}
		c = list[index]
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

func (d *Display) gotoContent(title string) bool {
	list := d.listContent()
	if index, ok := list.Find(title); ok {
		d.setContent(list[index])
		return true
	}
	return false
}

/// Requires d.mu.Lock to be held.
func (d *Display) pauseContent() {
	msg, err := ws_proto.MakeCommandMessage(ws_proto.PauseContentCommand, nil)
	if err != nil {
		log.Error().Err(err).Msg("make pause command")
		return
	}

	// We have d.mu.Lock, use broadcast instead of sendMessage.
	err = d.broadcast(*msg)
	if err != nil {
		log.Error().Err(err).Msg("send pause command")
	}
}

/// Requires d.mu.Lock to be held.
func (d *Display) resumeContent() {
	msg, err := ws_proto.MakeCommandMessage(ws_proto.ResumeContentCommand, nil)
	if err != nil {
		log.Error().Err(err).Msg("make resume command")
		return
	}

	// We have d.mu.Lock, use broadcast instead of sendMessage.
	err = d.broadcast(*msg)
	if err != nil {
		log.Error().Err(err).Msg("send resume command")
	}
}
