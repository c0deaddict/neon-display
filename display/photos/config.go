package photos

type Config struct {
	AlbumsPath string `json:"albums_path,omitempty"`
	CachePath  string `json:"cache_path,omitempty"`
}
