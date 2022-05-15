package files

import (
        "bufio"
        "os"
        "github.com/pkg/errors"
)

const (
	ARTISTS_FILE_DEFAULT = "./data/artists.txt"
)

type ArtistsList []string

func (a *ArtistsList) Read(filename string) error {
	artistsFile, err := os.Open(filename)
	if err != nil {
		return errors.Wrapf(err, "failed to open artists file %s", filename)
	}
	scanner := bufio.NewScanner(artistsFile)
	for scanner.Scan() {
		artist := scanner.Text()
		*a = append(*a, artist)
	}
	if err := scanner.Err(); err != nil {
		artistsFile.Close()
		return errors.Wrapf(err,"failed to read artists file %s", filename)
	}
	if err := artistsFile.Close(); err != nil {
		return errors.Wrapf(err, "failed to close artists file %s", filename)
	}
	return nil
}
