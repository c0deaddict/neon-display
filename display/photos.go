package display

import (
	"fmt"
	"os"
	"path"

	"github.com/c0deaddict/neon-display/display/ws_proto"
)

type PhotoAlbum struct {
	title string
	path  string
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
		// TODO: filter on file extension?
		if file.Type().IsRegular() {
			// filepath := path.Join(p.path, file.Name())
			// start := time.Now()
			// _, err := exif.ReadTool(filepath)
			// if err != nil {
			// 	log.Error().Err(err).Msg("read exif")
			// } else {
			// 	elapsed := time.Since(start)
			// 	log.Info().Msgf("exif parse success: %s", elapsed)
			// }

			// TODO: read exif data for Caption and Date (maybe Location?)
			photos = append(photos, ws_proto.Photo{
				ImagePath: fmt.Sprintf("%s/%s", p.title, file.Name()),
				Caption:   file.Name(),
				Date:      "unknown",
			})
		}
	}

	return photos, nil
}

func (d *Display) readAlbums() ([]PhotoAlbum, error) {
	if d.config.PhotosPath == "" {
		return nil, nil
	}

	files, err := os.ReadDir(d.config.PhotosPath)
	if err != nil {
		return nil, err
	}

	albums := make([]PhotoAlbum, 0)
	for _, file := range files {
		if file.IsDir() {
			albums = append(albums, PhotoAlbum{
				title: file.Name(),
				path:  path.Join(d.config.PhotosPath, file.Name()),
			})
		}
	}

	return albums, nil
}
