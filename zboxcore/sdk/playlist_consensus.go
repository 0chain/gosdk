package sdk

import (
	"encoding/json"
	"sort"
)

type playlistConsensus struct {
	files       map[string]PlaylistFile
	consensuses map[string]*Consensus

	threshConsensus int
	fullConsensus   int
}

func createPlaylistConsensus(fullConsensus, threshConsensus int) *playlistConsensus {
	return &playlistConsensus{
		files:           make(map[string]PlaylistFile),
		consensuses:     make(map[string]*Consensus),
		threshConsensus: threshConsensus,
		fullConsensus:   fullConsensus,
	}
}

func (c *playlistConsensus) AddFile(body []byte) error {
	file := PlaylistFile{}

	if err := json.Unmarshal([]byte(body), &file); err != nil {
		return err
	}

	_, ok := c.files[file.LookupHash]

	if ok {
		c.consensuses[file.LookupHash].Done()
	} else {
		cons := &Consensus{}

		cons.Init(c.threshConsensus, c.fullConsensus)
		cons.Done()

		c.consensuses[file.LookupHash] = cons
		c.files[file.LookupHash] = file
	}

	return nil
}

func (c *playlistConsensus) AddFiles(body []byte) error {
	var files []PlaylistFile

	if err := json.Unmarshal([]byte(body), &files); err != nil {
		return err
	}

	for _, f := range files {
		_, ok := c.files[f.LookupHash]

		if ok {
			c.consensuses[f.LookupHash].Done()
		} else {
			cons := &Consensus{}

			cons.Init(c.threshConsensus, c.fullConsensus)
			cons.Done()

			c.consensuses[f.LookupHash] = cons
			c.files[f.LookupHash] = f
		}

	}

	return nil

}

func (c *playlistConsensus) GetConsensusResult() []PlaylistFile {

	files := make([]PlaylistFile, 0, len(c.files))

	for _, file := range c.files {
		cons := c.consensuses[file.LookupHash]

		if cons.isConsensusOk() {
			files = append(files, file)
		}
	}

	sort.Slice(files, func(i, j int) bool {
		l := files[i]
		r := files[j]

		if len(l.Name) < len(r.Name) {
			return true
		}

		if len(l.Name) > len(r.Name) {
			return false
		}

		return l.Name < r.Name
	})

	return files
}

type playlistFileConsensus struct {
	files       map[string]PlaylistFile
	consensuses map[string]*Consensus

	threshConsensus float32
	fullConsensus   float32
	consensusOK     float32
}
