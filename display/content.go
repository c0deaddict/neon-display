package display

import (
	"fmt"
	"sort"

	"github.com/c0deaddict/neon-display/display/photos"
	"github.com/c0deaddict/neon-display/display/ws_proto"
	"github.com/rs/zerolog/log"
)

type content interface {
	Title() string
	Order() int
	Show() (*ws_proto.ShowContent, error)
	String() string
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

	albums, err := photos.ReadAlbums(d.config.PhotosPath, d.config.CachePath)
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

func (d *Display) refreshContent() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.content = d.listContent()
	if d.content.Len() == 0 {
		return fmt.Errorf("no content found")
	}

	log.Info().Msg("found content:")
	for _, c := range d.content {
		log.Info().Msgf("%s", c)
	}

	selected := ""
	if d.current >= 0 {
		selected = d.content[d.current].Title()
	} else {
		selected = d.config.InitTitle
	}

	if selected != "" {
		if index, ok := d.content.Find(selected); ok {
			d.current = index
			log.Info().Msgf("selected content: %s", d.content[index])
		} else {
			log.Warn().Msgf("selected content not found: %s", selected)
		}
	}

	if d.current < 0 {
		d.current = 0
	}

	return nil
}

func (d *Display) showContent() {
	show, err := d.content[d.current].Show()
	if err != nil {
		log.Error().Err(err).Msg("show content")
		return
	}

	msg, err := ws_proto.MakeCommandMessage(ws_proto.ShowContentCommand, show)
	if err != nil {
		log.Error().Err(err).Msg("make show command message")
		return
	}

	err = d.sendMessage(*msg)
	if err != nil {
		log.Error().Err(err).Msg("send show content command")
	}
}

func (d *Display) contentStep(step int) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if len(d.content) == 0 {
		log.Error().Msg("no content")
		return
	}

	index := (d.current + step) % d.content.Len()
	// Go's modulo can return negative numbers.
	if index < 0 {
		index = d.content.Len() + index
	}
	if index != d.current {
		d.current = index
		d.showContent()
	}
}

func (d *Display) prevContent() {
	d.contentStep(-1)
}

func (d *Display) nextContent() {
	d.contentStep(1)
}

func (d *Display) gotoContent(title string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	if index, ok := d.content.Find(title); ok {
		d.current = index
		// TODO: when content is a photoalbum this call can take a loooong time.
		d.showContent()
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

	err = d.sendMessage(*msg)
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

	err = d.sendMessage(*msg)
	if err != nil {
		log.Error().Err(err).Msg("send resume command")
	}
}
