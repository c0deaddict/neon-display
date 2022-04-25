package display

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/c0deaddict/neon-display/display/ws_proto"
)

type Video struct {
	title string
	path  string
}

func (v Video) Title() string {
	return v.title
}

func (v Video) Order() int {
	return 2000
}

func (v Video) Show() (*ws_proto.ShowContent, error) {
	data := ws_proto.VideoContent{
		Path: v.path,
	}

	return ws_proto.MakeShowContentMessage(ws_proto.VideoContentType, v.title, data)
}

func (d *Display) readVideos() ([]Video, error) {
	if d.config.VideosPath == "" {
		return nil, nil
	}

	files, err := os.ReadDir(d.config.VideosPath)
	if err != nil {
		return nil, err
	}

	videos := make([]Video, 0)
	for _, file := range files {
		if file.Type().IsRegular() {
			ext := filepath.Ext(file.Name())
			title := strings.TrimSuffix(file.Name(), ext)
			videos = append(videos, Video{
				title: title,
				path:  file.Name(),
			})
		}
	}

	return videos, nil
}
