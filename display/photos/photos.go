package photos

import (
	"fmt"
	"os"
	"path"

	"github.com/c0deaddict/neon-display/display/ws_proto"
	"github.com/rs/zerolog/log"
)

type PhotoAlbum struct {
	title  string
	path   string
	config Config
}

func (p PhotoAlbum) String() string {
	return fmt.Sprintf("PhotoAlbum %s", p.title)
}

func (p PhotoAlbum) Title() string {
	return p.title
}

func (p PhotoAlbum) Order() int {
	return 1000
}

func (p PhotoAlbum) Show() (*ws_proto.ShowContent, error) {
	photos, err := p.readPhotos()
	if err != nil {
		return nil, fmt.Errorf("read photos: %v", err)
	}

	data := ws_proto.PhotosContent{
		DelaySeconds: 10,
		Photos:       photos,
	}

	return ws_proto.MakeShowContentMessage(ws_proto.PhotosContentType, p.title, data)
}

func (p PhotoAlbum) readPhotos() ([]ws_proto.Photo, error) {
	files, err := os.ReadDir(p.path)
	if err != nil {
		return nil, err
	}

	photos := make([]ws_proto.Photo, 0)
	for _, file := range files {
		if file.Type().IsRegular() {
			filename := path.Join(p.path, file.Name())
			tags, err := readExif(filename, p.config.CachePath)
			if err != nil {
				log.Error().Err(err).Msg("read exif")
			}

			p := ws_proto.Photo{
				ImagePath: fmt.Sprintf("%s/%s", p.title, file.Name()),
			}

			if tags != nil {
				if value, ok := tags["DateTimeOriginal"]; ok {
					dt := value.(string)
					p.DateTime = &dt
				}

				if value, ok := tags["UserComment"]; ok {
					s := value.(string)
					p.Description = &s
				}

				if value, ok := tags["Model"]; ok {
					s := value.(string)
					p.Camera = &s
				}
			}

			photos = append(photos, p)
		}
	}

	return photos, nil
}

func ReadAlbums(cfg Config) ([]PhotoAlbum, error) {
	if cfg.AlbumsPath == "" {
		return nil, nil
	}

	files, err := os.ReadDir(cfg.AlbumsPath)
	if err != nil {
		return nil, err
	}

	albums := make([]PhotoAlbum, 0)
	for _, file := range files {
		if file.IsDir() {
			albums = append(albums, PhotoAlbum{
				title:  file.Name(),
				path:   path.Join(cfg.AlbumsPath, file.Name()),
				config: cfg,
			})
		}
	}

	return albums, nil
}
