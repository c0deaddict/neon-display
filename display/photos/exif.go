package photos

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"time"

	"github.com/rs/zerolog/log"
)

func readExif(filename string, cachePath string) (map[string]interface{}, error) {
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
		var tags map[string]interface{}
		err = json.NewDecoder(cf).Decode(&tags)
		if err != nil {
			return nil, fmt.Errorf("decode cache file: %v", err)
		}
		return tags, nil
	}

	start := time.Now()

	contents, err := exec.Command("exiftool", "-json", "-fast2", filename).Output()
	if err != nil {
		return nil, err
	}

	var res []map[string]interface{}
	err = json.Unmarshal(contents, &res)
	if err != nil {
		return nil, err
	}

	tags := res[0]

	elapsed := time.Since(start)
	log.Info().Dur("duration", elapsed).Str("filename", filename).Msg("read exif")

	// Write to cache.
	err = os.MkdirAll(filepath.Dir(cacheFile), os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("cache file mkdirs: %v", err)
	}
	cf, err := os.Create(cacheFile)
	if err != nil {
		return nil, fmt.Errorf("cache file create: %v", err)
	}
	defer cf.Close()
	err = json.NewEncoder(cf).Encode(tags)
	if err != nil {
		return nil, fmt.Errorf("encode cache file: %v", err)
	}

	return tags, nil
}
