package utils

import (
	"bufio"
	"cmd-notes/note"
	"encoding/gob"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// ReadState reads the current state of the notes from
// the given path.
func ReadState(path string) []note.Note {
	notes := make([]note.Note, 0)
	stateFilePath := filepath.Join(path, "state")

	file, err := os.Open(stateFilePath)
	if os.IsNotExist(err) {
		return notes
	} else if err != nil {
		log.Fatal("error opening state file: ", err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	dec := gob.NewDecoder(reader)

	if err := dec.Decode(&notes); err != nil {
		log.Fatal("failed to decode file: ", err)
	}

	return notes
}

// WriteState writes the current state of the notes to
// the given path. It first writes the state to a temporary
// file and then renames it to prevent corruption of
// the notes.
func WriteState(path string, notes *[]note.Note) {
	tempFilePath := filepath.Join(path, "state.temp")
	stateFilePath := filepath.Join(path, "state")

	slices.SortFunc(*notes, func(a, b note.Note) int {
		if a.Priority != b.Priority {
			return b.Priority - a.Priority
		}

		return strings.Compare(a.Contents, b.Contents)
	})

	err := os.MkdirAll(path, 0755)
	if err != nil {
		log.Fatal("error creating state directories: ", err)
	}

	file, err := os.Create(tempFilePath)
	if err != nil {
		log.Fatal("error creating temporary state file: ", err)
	}

	writer := bufio.NewWriter(file)

	enc := gob.NewEncoder(writer)
	if err := enc.Encode(notes); err != nil {
		log.Fatal("error encoding data: ", err)
	}

	writer.Flush()

	file.Close()

	err = os.Rename(file.Name(), stateFilePath)
	if err != nil {
		log.Fatal("error renaming state file: ", err)
	}
}
