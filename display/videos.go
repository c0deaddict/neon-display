package display

import (
	"os"

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
		Title: v.title,
		Path:  v.path,
	}

	return ws_proto.MakeShowContentMessage(ws_proto.VideoContentType, data)
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
			videos = append(videos, Video{
				title: file.Name(),
				path:  file.Name(),
			})
		}
	}

	return videos, nil
}
