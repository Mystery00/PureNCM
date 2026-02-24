package ncm

import "encoding/json"

// Meta holds the decoded song metadata from the NCM file's metadata block.
// Note: musicId and albumId are intentionally omitted — they are internal
// NetEase database keys with no use in conversion, and their JSON type
// varies across NCM versions (string vs number), making them error-prone.
type Meta struct {
	MusicName string   `json:"musicName"`
	Artist    [][2]any `json:"artist"` // [[name, id], ...]
	Album     string   `json:"album"`
	AlbumPic  string   `json:"albumPic"` // cover art URL
	Bitrate   int      `json:"bitrate"`
	Format    string   `json:"format"` // "mp3" or "flac"
}

// Artists returns a comma-joined string of artist names.
func (m *Meta) Artists() string {
	names := make([]string, 0, len(m.Artist))
	for _, a := range m.Artist {
		if len(a) > 0 {
			if name, ok := a[0].(string); ok {
				names = append(names, name)
			}
		}
	}
	result := ""
	for i, n := range names {
		if i > 0 {
			result += "/"
		}
		result += n
	}
	return result
}

// parseMeta decodes JSON metadata from the decrypted meta block.
// The raw bytes start with "music:" prefix before the JSON payload.
func parseMeta(raw []byte) (*Meta, error) {
	const prefix = "music:"
	if len(raw) > len(prefix) && string(raw[:len(prefix)]) == prefix {
		raw = raw[len(prefix):]
	}
	var m Meta
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, err
	}
	return &m, nil
}
