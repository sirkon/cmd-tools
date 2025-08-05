package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/alecthomas/kong"
	"github.com/dsoprea/go-exif/v3"
	"github.com/sirkon/errors"
	"github.com/sirkon/message"
)

const (
	exportedDir = "EXPORTED"
)

func main() {
	var cli struct {
		Move bool   `short:"m" help:"Move files instead of copying." default:"false"`
		Src  string `arg:"" help:"Path to a source directory with image files." default:"."`
	}

	parser := kong.Must(
		&cli,
		kong.Name("gallery"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Summary: true,
			Compact: true,
		}),
	)
	_, err := parser.Parse(os.Args[1:])
	if err != nil {
		parser.FatalIfErrorf(err)
	}

	action := "copy"
	if cli.Move {
		action = "move"
	}

	if err := copyToGallery(cli.Src, cli.Move); err != nil {
		message.Fatal(
			errors.Wrapf(err, "%s to gallery folder from source", action).Str("source-path", cli.Src),
		)
	}
}

func copyToGallery(src string, move bool) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return errors.Wrap(err, "get home dir")
	}

	gallerySuffix := []string{"Gallery", "Галерея"}
	var gallery string
	for _, path := range gallerySuffix {
		gal := filepath.Join(home, "Pictures", path)
		if _, err := os.Stat(gal); err == nil {
			gallery = gal
			break
		}
	}

	if gallery == "" {
		return errors.New("gallery root not found in ~/Pictures").Strs("possible-names", gallerySuffix)
	}

	dir, err := os.ReadDir(src)
	if err != nil {
		return errors.Wrap(err, "list source directory")
	}

	mkDirCache := make(map[string]struct{})
	if err := os.MkdirAll(filepath.Join(src, exportedDir), 0755); err != nil {
		return errors.Wrapf(err, "create %s directory", exportedDir)
	}

	for _, entry := range dir {
		if entry.IsDir() {
			continue
		}

		if err := processFile(entry, src, gallery, move, mkDirCache); err != nil {
			message.Error(errors.Wrap(err, "process file").Str("file", entry.Name()))
		}
	}

	return nil
}

func processFile(entry os.DirEntry, src string, gallery string, move bool, cache map[string]struct{}) error {
	sourceFile := filepath.Join(src, entry.Name())
	data, err := os.ReadFile(sourceFile)
	if err != nil {
		return errors.Wrap(err, "read file")
	}

	rawExif, err := exif.SearchAndExtractExif(data)
	if err != nil {
		return errors.Wrap(err, "search and extract raw exif data")
	}

	tags, _, err := exif.GetFlatExifDataUniversalSearch(rawExif, &exif.ScanOptions{}, true)
	if err != nil {
		return errors.Wrap(err, "parse exif data")
	}

	for _, tag := range tags {
		if tag.TagId != 0x0132 {
			continue
		}

		const exifDateTimeTagName = "DateTime"
		if tag.TagName != exifDateTimeTagName {
			return errors.Newf("unexpected tag name %q", tag.TagName).
				Uint16("tag-id", tag.TagId).
				Str("expected-tag-name", exifDateTimeTagName)
		}

		const exifStringTypeName = "ASCII"
		if tag.TagTypeName != exifStringTypeName {
			return errors.Newf("unexpected tag type %q for the datetime", tag.TagTypeName).
				Uint16("tag-id", tag.TagId).
				Str("expected-tag-type", exifStringTypeName)
		}

		dateTimeTagValue := tag.Value.(string)
		datetime, err := time.Parse("2006:01:02 15:04:05", dateTimeTagValue)
		if err != nil {
			return errors.Wrap(err, "parse file datetime tag").Str("invalid-datetime", dateTimeTagValue)
		}

		year := strconv.Itoa(datetime.Year())
		month := fmt.Sprintf("%02d", datetime.Month())
		day := fmt.Sprintf("%02d", datetime.Day())
		dir := filepath.Join(gallery, year, month, day)
		if _, ok := cache[dir]; !ok {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return errors.Wrap(err, "create daily directory").Str("date", datetime.Format(time.DateOnly))
			}

			cache[dir] = struct{}{}
		}

		destFile := filepath.Join(dir, entry.Name())
		if move {
			if err := os.Rename(sourceFile, destFile); err != nil {
				if err := os.RemoveAll(destFile); err != nil {
					message.Warning(errors.Wrapf(err, "delete destination file %q after move failure", destFile))
				}

				return errors.Wrapf(err, "move file %q -> %q", sourceFile, destFile)
			}
		} else {
			_, err := os.Stat(destFile)
			exportedDest := filepath.Join(src, exportedDir, entry.Name())
			if err == nil {
				message.Warningf("file %q already exists", destFile)
				if err := os.Rename(sourceFile, exportedDest); err != nil {
					message.Warning(errors.Wrap(err, "move file to exported dir"))
				}
				return nil
			}
			if !os.IsNotExist(err) {
				return errors.Wrap(err, "check if destination file exists")
			}

			if err := copyFile(sourceFile, destFile); err != nil {
				return errors.Wrap(err, "copy file")
			}

			if err := os.Rename(sourceFile, exportedDest); err != nil {
				message.Warning(errors.Wrap(err, "move file to exported dir"))
			}
		}

		return nil
	}

	message.Warningf("no datetime tag found for %q", entry.Name())
	return nil
}

func copyFile(sourceFile string, destFile string) (err error) {
	message.Infof("copying %q -> %q", sourceFile, destFile)

	srcFile, err := os.Open(sourceFile)
	if err != nil {
		return errors.Wrap(err, "open source file")
	}
	defer func() {
		if err := srcFile.Close(); err != nil {
			message.Warning(errors.Wrapf(err, "close source file %q", sourceFile))
		}
	}()

	dstFile, err := os.Create(destFile)
	if err != nil {
		return errors.Wrapf(err, "create destination file %q", destFile)
	}
	defer func() {
		if err := dstFile.Close(); err != nil {
			message.Warning(errors.Wrapf(err, "close destination file %q", destFile))
		} else {
			return
		}

		if err := os.Remove(destFile); err != nil {
			message.Warning(errors.Wrapf(err, "remove destination file %q after the copy failure", destFile))
		}
	}()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return errors.Wrap(err, "move source file contents into the destination file").
			Str("destination", destFile).
			Str("source", sourceFile)
	}

	return nil
}
