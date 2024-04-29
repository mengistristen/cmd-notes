// This is an extemely simple package for creating and tracking
// short note snippets.
package main

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/spf13/cobra"
)

const (
	TODO = iota
	IN_PROGRESS
	COMPLETE
)

type Note struct {
	State    int
	Contents string
}

// Notes is a type alias for a simple string slice.
type Notes []Note

func main() {
	command := setup()

	if err := command.Execute(); err != nil {
		log.Fatal(err)
	}
}

// setup creates the environment for the command line
// parser.
func setup() *cobra.Command {
	rootCmd := &cobra.Command{
		Use: "cmd-notes",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			basePath := getDataBasePath()

			cmd.Root().Annotations["stateFilePath"] = basePath
		},
	}

	rootCmd.Annotations = make(map[string]string)

	cmdAdd := &cobra.Command{
		Use:   "add \"<note>\"",
		Short: "Add note",
		Run:   addNote,
	}

	cmdRemove := &cobra.Command{
		Use:   "rm <index>",
		Short: "Remove note",
		Run:   removeNote,
	}

	cmdList := &cobra.Command{
		Use:   "ls",
		Short: "List notes",
		Run:   listNotes,
	}

	rootCmd.AddCommand(cmdAdd, cmdList, cmdRemove)

	return rootCmd
}

// addNote creates a new note.
func addNote(cmd *cobra.Command, args []string) {
	path := cmd.Root().Annotations["stateFilePath"]
	notes := readState(path)

	if len(args) < 1 {
		log.Fatal("usage: cmd-notes add \"<note>\"")
	}

	notes = append(notes, Note{
		State:    TODO,
		Contents: args[0],
	})

	writeState(path, &notes)

	fmt.Println("\033[32m \u2713 \033[0madded note")
}

// removeNote removes an existing note.
func removeNote(cmd *cobra.Command, args []string) {
	path := cmd.Root().Annotations["stateFilePath"]
	notes := readState(path)

	if len(args) < 1 {
		log.Fatal("usage: cmd-notes rm <index>")
	}

	index, err := strconv.Atoi(args[0])
	if err != nil {
		log.Fatal("error parsing note index: ", err)
	}

	if index >= len(notes) {
		log.Fatal("invalid note index")
	}

	notes = append(notes[:index], notes[index+1:]...)

	writeState(path, &notes)

	fmt.Println("\033[31m \u2717 \033[0mremoved note")
}

// listNotes lists all existing notes.
func listNotes(cmd *cobra.Command, args []string) {
	path := cmd.Root().Annotations["stateFilePath"]
	notes := readState(path)

	for index, note := range notes {
		color := "\033[0m"

		switch note.State {
		case TODO:
			color = "\033[91m"
		case IN_PROGRESS:
			color = "\033[93m"
		case COMPLETE:
			color = "\033[92m"
		}

		fmt.Printf("\033[94m %d \033[0m- %s%s\n", index, color, note.Contents)
	}
}

// getDataBasePath finds the directory in which to place the
// current state of the notes.
func getDataBasePath() string {
	var basePath string
	var exists bool

	switch runtime.GOOS {
	case "windows":
		basePath, exists = os.LookupEnv("LOCALAPPDATA")
		if !exists {
			log.Fatal("failed to locate local app data")
		}
	case "darwin":
	case "linux":
		basePath, exists = os.LookupEnv("XDG_STATE_HOME")
		if !exists {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				log.Fatal("error locating user's home directory: ", err)
			}

			basePath = filepath.Join(homeDir, ".local/state")
		}
	}

	basePath = filepath.Join(basePath, "cmd-notes")

	return basePath
}

// readState reads the current state of the notes from
// the given path.
func readState(path string) Notes {
	notes := make([]Note, 0)
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

// writeState writes the current state of the notes to
// the given path. It first writes the state to a temporary
// file and then renames it to prevent corruption of
// the notes.
func writeState(path string, notes *Notes) {
	tempFilePath := filepath.Join(path, "state.temp")
	stateFilePath := filepath.Join(path, "state")

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
