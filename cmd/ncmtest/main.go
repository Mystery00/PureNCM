package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"PureNCM/internal/ncm"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: ncmtest <file.ncm> [output_dir]")
		os.Exit(1)
	}

	inputPath := os.Args[1]
	outputDir := filepath.Dir(inputPath) // default: same dir as input
	if len(os.Args) >= 3 {
		outputDir = os.Args[2]
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("cannot create output dir: %v", err)
	}

	log.Printf("Decrypting: %s", inputPath)
	result, err := ncm.DecryptFile(inputPath)
	if err != nil {
		log.Fatalf("decryption failed: %v", err)
	}
	if result.MetaErr != nil {
		log.Printf("WARNING: metadata parse failed: %v (audio will still be written without tags)", result.MetaErr)
	}

	log.Printf("Format   : %s", result.Format)
	log.Printf("Title    : %s", result.Meta.MusicName)
	log.Printf("Artist   : %s", result.Meta.Artists())
	log.Printf("Album    : %s", result.Meta.Album)
	log.Printf("Cover URL: %s", result.Meta.AlbumPic)
	log.Printf("Audio    : %d bytes", len(result.Audio))

	outPath, err := ncm.WriteToFile(result, outputDir)
	if err != nil {
		log.Fatalf("write failed: %v", err)
	}
	log.Printf("Output   : %s", outPath)
}
