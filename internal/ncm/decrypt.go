package ncm

import (
	"bytes"
	"crypto/aes"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"io"
	"os"
)

// NCM format constants
var (
	magicHeader = []byte{0x43, 0x54, 0x45, 0x4E, 0x46, 0x44, 0x41, 0x4D} // "CTENFDAM"
	coreKey     = []byte{0x68, 0x7A, 0x48, 0x52, 0x41, 0x6D, 0x73, 0x6F,
		0x35, 0x6B, 0x49, 0x6E, 0x62, 0x61, 0x78, 0x57} // NCM core AES key
	metaKey = []byte{0x23, 0x31, 0x34, 0x6C, 0x6A, 0x6B, 0x5F, 0x21,
		0x5C, 0x5D, 0x26, 0x30, 0x55, 0x3C, 0x27, 0x28} // NCM meta AES key
)

// DecryptResult holds the decrypted audio data and parsed metadata.
type DecryptResult struct {
	Meta      *Meta
	Audio     []byte // raw mp3 or flac bytes
	CoverData []byte // cover art bytes (embedded in NCM or nil)
	Format    string // "mp3" or "flac"
	// MetaErr is non-nil when metadata parsing partially failed.
	// Audio decryption is unaffected; Meta will contain zero values for
	// fields that could not be parsed.
	MetaErr error
}

// DecryptFile reads an NCM file and decrypts it, returning audio bytes and metadata.
func DecryptFile(path string) (*DecryptResult, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return Decrypt(f)
}

// Decrypt performs the full NCM decryption pipeline on the given reader.
func Decrypt(r io.Reader) (*DecryptResult, error) {
	// 1. Validate magic header
	magic := make([]byte, 8)
	if _, err := io.ReadFull(r, magic); err != nil {
		return nil, err
	}
	if !bytes.Equal(magic, magicHeader) {
		return nil, errors.New("not a valid NCM file: magic header mismatch")
	}

	// Skip 2 gap bytes
	if _, err := io.ReadFull(r, make([]byte, 2)); err != nil {
		return nil, err
	}

	// 2. Read & decrypt the RC4 key block (AES-128-ECB with coreKey)
	keyLen, err := readUint32LE(r)
	if err != nil {
		return nil, err
	}
	keyData := make([]byte, keyLen)
	if _, err := io.ReadFull(r, keyData); err != nil {
		return nil, err
	}
	// XOR each byte with 0x64
	for i := range keyData {
		keyData[i] ^= 0x64
	}
	decryptedKey, err := aesECBDecrypt(keyData, coreKey)
	if err != nil {
		return nil, err
	}
	// decryptedKey starts with "neteasecloudmusic" prefix — skip it
	const keyPrefix = "neteasecloudmusic"
	if len(decryptedKey) > len(keyPrefix) && string(decryptedKey[:len(keyPrefix)]) == keyPrefix {
		decryptedKey = decryptedKey[len(keyPrefix):]
	}
	rc4Key := decryptedKey

	// 3. Read & decrypt metadata block (AES-128-ECB with metaKey)
	metaLen, err := readUint32LE(r)
	if err != nil {
		return nil, err
	}
	var meta *Meta
	var coverData []byte
	if metaLen > 0 {
		metaData := make([]byte, metaLen)
		if _, err := io.ReadFull(r, metaData); err != nil {
			return nil, err
		}
		// XOR each byte with 0x63
		for i := range metaData {
			metaData[i] ^= 0x63
		}
		// Strip the "163 key(Don't modify):" header before base64
		const metaPrefix = "163 key(Don't modify):"
		if len(metaData) > len(metaPrefix) && string(metaData[:len(metaPrefix)]) == metaPrefix {
			metaData = metaData[len(metaPrefix):]
		}
		decoded, err := base64.StdEncoding.DecodeString(string(metaData))
		if err != nil {
			decoded, err = base64.RawStdEncoding.DecodeString(string(metaData))
			if err != nil {
				return nil, err
			}
		}
		metaDecrypted, err := aesECBDecrypt(decoded, metaKey)
		if err != nil {
			return nil, err
		}
		var metaErr error
		meta, metaErr = parseMeta(metaDecrypted)
		if metaErr != nil {
			// Non-fatal: audio decryption continues, but MetaErr is surfaced
			// to the caller so it can be logged or shown in the UI.
			meta = &Meta{}
			coverData = nil // can't trust cover either
			return &DecryptResult{Meta: meta, MetaErr: metaErr}, nil
		}
	} else {
		meta = &Meta{}
	}

	// 4. Skip CRC32 (4 bytes) + gap (5 bytes)
	if _, err := io.ReadFull(r, make([]byte, 9)); err != nil {
		return nil, err
	}

	// 5. Read & skip embedded cover image
	coverImgLen, err := readUint32LE(r)
	if err != nil {
		return nil, err
	}
	if coverImgLen > 0 {
		coverData = make([]byte, coverImgLen)
		if _, err := io.ReadFull(r, coverData); err != nil {
			return nil, err
		}
	}

	// 6. Build RC4 keystream (using the S-Box / KSA algorithm as NCM uses)
	keyBox := buildRC4KeyBox(rc4Key)

	// 7. Decrypt audio stream via XOR
	audio, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	for i := range audio {
		j := (i + 1) & 0xFF
		audio[i] ^= keyBox[(int(keyBox[j])+int(keyBox[(int(j)+int(keyBox[j]))&0xFF]))&0xFF]
	}

	// 8. Detect actual format from audio header
	format := detectFormat(audio)
	if meta.Format == "" {
		meta.Format = format
	}

	return &DecryptResult{
		Meta:      meta,
		Audio:     audio,
		CoverData: coverData,
		Format:    format,
	}, nil
}

// buildRC4KeyBox constructs the NCM-specific RC4 S-Box (KSA step).
func buildRC4KeyBox(key []byte) [256]byte {
	var box [256]byte
	for i := range box {
		box[i] = byte(i)
	}
	keyLen := len(key)
	var j byte
	for i := 0; i < 256; i++ {
		j = j + box[i] + key[i%keyLen]
		box[i], box[j] = box[j], box[i]
	}
	return box
}

// detectFormat inspects the first few bytes to determine if audio is MP3 or FLAC.
func detectFormat(data []byte) string {
	if len(data) >= 4 {
		// fLaC
		if data[0] == 0x66 && data[1] == 0x4C && data[2] == 0x61 && data[3] == 0x43 {
			return "flac"
		}
	}
	return "mp3"
}

// aesECBDecrypt decrypts data with AES-128-ECB and removes PKCS7 padding.
func aesECBDecrypt(data, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	bs := block.BlockSize()
	if len(data)%bs != 0 {
		return nil, errors.New("aes ecb: data length not a multiple of block size")
	}
	dst := make([]byte, len(data))
	for i := 0; i < len(data); i += bs {
		block.Decrypt(dst[i:], data[i:])
	}
	return pkcs7Unpad(dst)
}

// pkcs7Unpad removes PKCS7 padding.
func pkcs7Unpad(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("pkcs7: empty data")
	}
	pad := int(data[len(data)-1])
	if pad == 0 || pad > aes.BlockSize || pad > len(data) {
		return nil, errors.New("pkcs7: invalid padding")
	}
	return data[:len(data)-pad], nil
}

// readUint32LE reads a little-endian uint32 from r.
func readUint32LE(r io.Reader) (uint32, error) {
	buf := make([]byte, 4)
	if _, err := io.ReadFull(r, buf); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf), nil
}
