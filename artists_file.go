package main

import "github.com/pkg/errors"

type ArtistsList []string

func (a *ArtistsList) Read(filename string) error {
	artistsFile, err := os.Open(fileName)
	if err != nil {
		return errors.Wrapf("failed to open artists file %s", filename)
	}
	scanner := bufio.NewScanner(artistsFile)
	for scanner.Scan() {
		artist := scanner.Text()
		artists = append(artists, artist)
	}
	if err := scanner.Err(); err != nil {
		artistsFile.Close()
		return errors.Wrapf("failed to read artists file %s", filename)
	}
	if err := artistsFile.Close(); err != nil {
		return errors.Wrapf("failed to close artists file %s", filename)
	}
	return nil
}
