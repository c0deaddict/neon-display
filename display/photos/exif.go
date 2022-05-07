package photos

import (
	"crypto/sha256"
	"encoding/gob"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"

	goexif "github.com/dsoprea/go-exif/v3"
	"github.com/rs/zerolog/log"
)

func readExif(filename string, cachePath string) ([]goexif.ExifTag, error) {
	fileInfo, err := os.Stat(filename)
	if err != nil {
		return nil, err
	}

	hash := sha256.Sum256([]byte(filename))
	cacheFile := path.Join(
		cachePath,
		fmt.Sprintf("%02x", hash[0:1]),
		fmt.Sprintf("%015x", hash[1:16]),
	)

	cacheInfo, err := os.Stat(cacheFile)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	// Cache hit.
	if err == nil && !fileInfo.ModTime().After(cacheInfo.ModTime()) {
		cf, err := os.Open(cacheFile)
		if err != nil {
			return nil, fmt.Errorf("open cache file: %v", err)
		}
		defer cf.Close()
		var res []goexif.ExifTag
		err = gob.NewDecoder(cf).Decode(&res)
		if err != nil {
			return nil, fmt.Errorf("decode cache file: %v", err)
		}
		return res, nil
	}

	start := time.Now()

	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	data, err := goexif.SearchAndExtractExifWithReader(f)
	if err != nil {
		return nil, err
	}

	tags, _, err := goexif.GetFlatExifData(data, nil)
	if err != nil {
		return nil, err
	}

	elapsed := time.Since(start)
	log.Info().Dur("duration", elapsed).Str("filename", filename).Msg("read exif")

	// Write to cache.
	err = os.MkdirAll(filepath.Dir(cacheFile), os.ModePerm)
	if err != nil {
		return nil, err
	}
	cf, err := os.Create(cacheFile)
	if err != nil {
		return nil, err
	}
	defer cf.Close()
	gob.NewEncoder(cf).Encode(tags)

	return tags, nil
}
