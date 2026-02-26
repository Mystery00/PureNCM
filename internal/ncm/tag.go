package ncm

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	id3v2 "github.com/bogem/id3v2/v2"
	flacvorbis "github.com/go-flac/flacvorbis"
	flac "github.com/go-flac/go-flac"
)

// WriteToFile writes the decrypted audio with embedded tags to the output file.
// filenamePattern supports {title}, {artist}, {album} placeholders.
func WriteToFile(result *DecryptResult, outputDir string, filenamePattern string) (string, error) {
	return WriteToFileWithProgress(result, outputDir, filenamePattern, nil)
}

// WriteToFileWithProgress is like WriteToFile but calls progressFn(0..1) during the write.
// progressFn may be nil.
func WriteToFileWithProgress(result *DecryptResult, outputDir string, filenamePattern string, progressFn func(float64)) (string, error) {
	meta := result.Meta
	cover := result.CoverData

	// Download cover art if not embedded in the NCM file
	if len(cover) == 0 && meta.AlbumPic != "" {
		cover, _ = downloadCover(meta.AlbumPic)
	}

	// Build output filename from pattern
	name := applyPattern(filenamePattern, meta)
	if name == "" {
		name = sanitizeFilename(meta.MusicName)
	}
	if name == "" {
		name = fmt.Sprintf("track_%d", time.Now().Unix())
	}
	ext := "." + result.Format
	outPath := filepath.Join(outputDir, name+ext)

	switch result.Format {
	case "flac":
		return outPath, writeFlacTags(result.Audio, outPath, meta, cover, progressFn)
	default: // mp3
		return outPath, writeMp3Tags(result.Audio, outPath, meta, cover, progressFn)
	}
}

// countingWriter wraps an io.Writer and calls onWrite with cumulative progress (0..1).
type countingWriter struct {
	w       io.Writer
	total   int64
	written int64
	fn      func(float64)
}

func (cw *countingWriter) Write(p []byte) (n int, err error) {
	n, err = cw.w.Write(p)
	if n > 0 && cw.fn != nil && cw.total > 0 {
		cw.written += int64(n)
		pct := float64(cw.written) / float64(cw.total)
		if pct > 1 { pct = 1 }
		cw.fn(pct)
	}
	return
}

// applyPattern replaces {title}, {artist}, {album} in pattern and sanitizes the result.
func applyPattern(pattern string, meta *Meta) string {
	if pattern == "" {
		pattern = "{title}"
	}
	result := pattern
	result = strings.ReplaceAll(result, "{title}", meta.MusicName)
	result = strings.ReplaceAll(result, "{artist}", meta.Artists())
	result = strings.ReplaceAll(result, "{album}", meta.Album)
	return sanitizeFilename(result)
}

// writeMp3Tags writes audio bytes + ID3v2 tags to an mp3 file.
func writeMp3Tags(audio []byte, path string, meta *Meta, cover []byte, progressFn func(float64)) error {
	// Write raw audio via countingWriter so we can report progress
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	cw := &countingWriter{w: f, total: int64(len(audio)), fn: progressFn}
	_, err = io.Copy(cw, bytes.NewReader(audio))
	f.Close()
	if err != nil {
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
func writeFlacTags(audio []byte, path string, meta *Meta, cover []byte, progressFn func(float64)) error {
	f, err := flac.ParseBytes(bytes.NewReader(audio))
	if err != nil {
		// If parse fails, write raw and return
		return writeWithProgress(path, audio, progressFn)
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
		return writeWithProgress(path, audio, progressFn)
	}
	if progressFn != nil {
		progressFn(1.0)
	}
	return nil
}

// writeWithProgress writes bytes to path, reporting progress via progressFn.
func writeWithProgress(path string, data []byte, progressFn func(float64)) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	cw := &countingWriter{w: f, total: int64(len(data)), fn: progressFn}
	_, err = io.Copy(cw, bytes.NewReader(data))
	return err
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
