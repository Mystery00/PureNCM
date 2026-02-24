package ncm

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	id3v2 "github.com/bogem/id3v2/v2"
	flacvorbis "github.com/go-flac/flacvorbis"
	flac "github.com/go-flac/go-flac"
)

// WriteToFile writes the decrypted audio with embedded tags to the output file.
// It downloads cover art from meta.AlbumPic if coverData is empty.
func WriteToFile(result *DecryptResult, outputDir string) (string, error) {
	meta := result.Meta
	cover := result.CoverData

	// Download cover art if not embedded in the NCM file
	if len(cover) == 0 && meta.AlbumPic != "" {
		cover, _ = downloadCover(meta.AlbumPic)
	}

	// Build a safe output filename
	name := meta.MusicName
	if name == "" {
		name = "unknown"
	}
	ext := "." + result.Format
	outPath := filepath.Join(outputDir, sanitizeFilename(name)+ext)

	switch result.Format {
	case "flac":
		return outPath, writeFlacTags(result.Audio, outPath, meta, cover)
	default: // mp3
		return outPath, writeMp3Tags(result.Audio, outPath, meta, cover)
	}
}

// writeMp3Tags writes audio bytes + ID3v2 tags to an mp3 file.
func writeMp3Tags(audio []byte, path string, meta *Meta, cover []byte) error {
	// Write raw audio first so bogem can open it
	if err := os.WriteFile(path, audio, 0644); err != nil {
		return err
	}
	tag, err := id3v2.Open(path, id3v2.Options{Parse: false})
	if err != nil {
		return err
	}
	defer tag.Close()

	tag.SetTitle(meta.MusicName)
	tag.SetArtist(meta.Artists())
	tag.SetAlbum(meta.Album)

	if len(cover) > 0 {
		picFrame := id3v2.PictureFrame{
			Encoding:    id3v2.EncodingUTF8,
			MimeType:    "image/jpeg",
			PictureType: id3v2.PTFrontCover,
			Description: "Cover",
			Picture:     cover,
		}
		tag.AddAttachedPicture(picFrame)
	}
	return tag.Save()
}

// writeFlacTags writes audio bytes + Vorbis Comment tags to a flac file.
func writeFlacTags(audio []byte, path string, meta *Meta, cover []byte) error {
	f, err := flac.ParseBytes(bytes.NewReader(audio))
	if err != nil {
		// If parse fails, write raw and return
		return os.WriteFile(path, audio, 0644)
	}

	// Build vorbis comment block
	cmt := flacvorbis.New()
	if meta.MusicName != "" {
		_ = cmt.Add(flacvorbis.FIELD_TITLE, meta.MusicName)
	}
	if a := meta.Artists(); a != "" {
		_ = cmt.Add(flacvorbis.FIELD_ARTIST, a)
	}
	if meta.Album != "" {
		_ = cmt.Add(flacvorbis.FIELD_ALBUM, meta.Album)
	}

	cmtBlock := cmt.Marshal()
	// Replace or append vorbis comment block (type 4)
	replaced := false
	for i, m := range f.Meta {
		if m.Type == flac.VorbisComment {
			f.Meta[i] = &cmtBlock
			replaced = true
			break
		}
	}
	if !replaced {
		f.Meta = append(f.Meta, &cmtBlock)
	}

	// Embed cover as PICTURE block if available
	if len(cover) > 0 {
		picBlock := buildFlacPictureBlock(cover)
		f.Meta = append(f.Meta, picBlock)
	}

	if err := f.Save(path); err != nil {
		// Fallback: write raw audio
		return os.WriteFile(path, audio, 0644)
	}
	return nil
}

// buildFlacPictureBlock creates a minimal PICTURE metadata block for FLAC.
func buildFlacPictureBlock(cover []byte) *flac.MetaDataBlock {
	// PICTURE block: type 3 (front cover), MIME "image/jpeg"
	// Binary layout: uint32 pic_type | uint32 mime_len | mime | uint32 desc_len | desc |
	//                uint32 width | uint32 height | uint32 color_depth | uint32 color_count |
	//                uint32 data_len | data
	mime := []byte("image/jpeg")
	desc := []byte{}
	var buf bytes.Buffer
	writeUint32BE := func(v uint32) {
		buf.Write([]byte{byte(v >> 24), byte(v >> 16), byte(v >> 8), byte(v)})
	}
	writeUint32BE(3) // front cover
	writeUint32BE(uint32(len(mime)))
	buf.Write(mime)
	writeUint32BE(uint32(len(desc)))
	buf.Write(desc)
	writeUint32BE(0) // width unknown
	writeUint32BE(0) // height unknown
	writeUint32BE(0) // color depth unknown
	writeUint32BE(0) // color count
	writeUint32BE(uint32(len(cover)))
	buf.Write(cover)

	return &flac.MetaDataBlock{
		Type: flac.Picture,
		Data: buf.Bytes(),
	}
}

// downloadCover fetches cover art from the given URL with a 10s timeout.
func downloadCover(url string) ([]byte, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// sanitizeFilename removes characters that are illegal in Windows filenames.
func sanitizeFilename(name string) string {
	illegal := `\/:*?"<>|`
	result := make([]rune, 0, len(name))
	for _, r := range name {
		skip := false
		for _, c := range illegal {
			if r == c {
				skip = true
				break
			}
		}
		if !skip {
			result = append(result, r)
		}
	}
	if len(result) == 0 {
		return fmt.Sprintf("track_%d", time.Now().Unix())
	}
	return string(result)
}
