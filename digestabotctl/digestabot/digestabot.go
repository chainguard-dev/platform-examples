package digestabot

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var regex = regexp.MustCompile(`[a-z0-9]+([._-][a-z0-9]+)*(/[a-z0-9]+([._-][a-z0-9]+)*)*@sha256:[a-z0-9]+`)
var shaRegex = regexp.MustCompile(`sha256:[a-z0-9]+`)
var DefaultFileTypes = []string{`*.yaml`, `*.yml`, `*.sh`, `*.tf`, `*.tfvars`, `Dockerfile*`, `Makefile*`}

type Image struct {
	Name        string
	CurrentHash string
	UpdatedHash string
}

func NewImageFromString(line string) *Image {
	return &Image{
		Name:        getImageName(line),
		CurrentHash: shaRegex.FindString(line),
	}
}

func getImageName(line string) string {
	shortened := line
	separators := map[string]int{
		"@": 0,
		"=": 0,
		" ": 1,
	}

	for sep, count := range separators {
		shortened = strings.SplitN(shortened, sep, 2)[count]
	}

	shortened = strings.TrimPrefix(shortened, "docker://")
	shortened = strings.TrimSpace(strings.ReplaceAll(shortened, "\t", ""))

	return shortened
}

type UpdateOptions struct {
	Name     string
	Digester Digester
	InFile   io.Reader
	OutFile  *bufio.Writer
	Logger   *slog.Logger
}

func UpdateHashes(opts UpdateOptions) error {
	scanner := bufio.NewScanner(opts.InFile)

	for scanner.Scan() {
		line := scanner.Bytes()
		newLine := string(line)

		if regex.Match(line) {
			image := NewImageFromString(string(line))
			hash := image.CurrentHash
			updated, err := opts.Digester.Digest(image.Name)
			if err != nil {
				return err
			}
			image.UpdatedHash = updated

			if image.CurrentHash != image.UpdatedHash {
				opts.Logger.Info("old", "hash", hash)
				opts.Logger.Info("new", "hash", updated)
				newLine = strings.Replace(string(line), hash, updated, 1)
			}
		}
		fmt.Fprintln(opts.OutFile, newLine)
	}

	return opts.OutFile.Flush()
}

func UpdateFiles(files []string, logger *slog.Logger) error {
	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			return err
		}
		defer f.Close()
		outFile := fmt.Sprintf("%s.tmp", file)
		out, err := os.Create(outFile)
		if err != nil {
			return err
		}
		defer out.Close()
		writer := bufio.NewWriter(out)

		opts := UpdateOptions{
			Name:     file,
			Digester: Crane{},
			InFile:   f,
			OutFile:  writer,
			Logger:   logger,
		}

		logger.Info("processing", "file", opts.Name)
		if err := UpdateHashes(opts); err != nil {
			return err
		}
		if err := os.Rename(outFile, file); err != nil {
			return err
		}

	}

	return nil
}

func FindFiles(fileTypes []string, directory string) ([]string, error) {
	files := []string{}

	if err := filepath.WalkDir(directory, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		base := filepath.Base(path)
		for _, pattern := range fileTypes {
			matched, err := filepath.Match(pattern, base)
			if err != nil {
				return err
			}
			if matched {
				files = append(files, path)
				break
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return files, nil

}
