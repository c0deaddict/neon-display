package exif

import (
	"encoding/json"
	"os"
	"os/exec"

	goexif "github.com/dsoprea/go-exif/v3"
)

func Read(filepath string) ([]goexif.ExifTag, error) {
	f, err := os.Open(filepath)
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

	return tags, nil
}

func ReadTool(filepath string) ([]map[string]interface{}, error) {
	contents, err := exec.Command("exiftool", "-json", "-fast2", filepath).Output()
	if err != nil {
		return nil, err
	}

	var tags []map[string]interface{}
	err = json.Unmarshal(contents, &tags)
	if err != nil {
		return nil, err
	}

	return tags, nil
}
