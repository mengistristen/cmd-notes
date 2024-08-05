// This is an extemely simple package for creating and tracking
// short note snippets.
package main

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
)

// Available states
const (
	NONE = iota
	TODO
	IN_PROGRESS
	COMPLETE
)

// Available priorities
const (
	LOW = iota
	MEDIUM
	HIGH
)

// ANSI escape codes for colors
const (
	red    = "\033[91m"
	yellow = "\033[93m"
	green  = "\033[92m"
	blue   = "\033[34m"
	reset  = "\033[0m"
)

var funcMap = template.FuncMap{
	"colorizeContents": func(state int, text string) string {
		var result string

		switch state {
		case TODO:
			result = red + text + reset
		case IN_PROGRESS:
			result = yellow + text + reset
		case COMPLETE:
			result = green + text + reset
		default:
			result = text
		}

		return result
	},
	"priority": func(priority int) string {
		var result string

		switch priority {
		case LOW:
			result = "(" + blue + "\u2193" + reset + ")"
		case MEDIUM:
			result = "(-)"
		case HIGH:
			result = "(" + red + "\u2191" + reset + ")"
		default:
			result = ""
		}

		return result
	},
	"status": func(status int) string {
		var result string

		switch status {
		case TODO:
			result = "Todo"
		case IN_PROGRESS:
			result = "In Progress"
		case COMPLETE:
			result = "Complete"
		default:
			result = ""
		}
		return result
	},
	"index": func(index int) string {
		return blue + strconv.Itoa(index) + reset
	},
	"filterByPriority": func(notes []Note, priority int) []Note {
		var filteredItems []Note

		for _, note := range notes {
			if note.Priority == priority {
				filteredItems = append(filteredItems, note)
			}
		}

		return filteredItems
	},
}

var templ string

type Formatter interface {
	format(w io.Writer, notes []Note)
}

type Note struct {
	Priority int
	State    int
	Contents string
}

type TerminalFormatter struct{}

func (f TerminalFormatter) format(w io.Writer, notes []Note) {
	for index, note := range notes {
		color := "\033[0m"
		priority := ""

		switch note.Priority {
		case LOW:
			priority = "(\033[34m\u2193\033[0m)"
		case MEDIUM:
			priority = "(-)"
		case HIGH:
			priority = "(\033[31m\u2191\033[0m)"
		}

		switch note.State {
		case TODO:
			color = "\033[91m"
		case IN_PROGRESS:
			color = "\033[93m"
		case COMPLETE:
			color = "\033[92m"
		}

		fmt.Fprintf(w, "%s\033[94m %d \033[0m- %s%s\033[0m\n", priority, index, color, note.Contents)
	}
}

type TemplateFormatter struct{}

func (f TemplateFormatter) format(w io.Writer, notes []Note) {
	path := getDataBasePath()
	name := templ + ".tmpl"

	t, err := template.New(name).Funcs(funcMap).ParseFiles(fmt.Sprintf("%s/templates/%s", path, name))
	if err != nil {
		log.Fatalf("Error parsing template file: %v", err)
	}

	err = t.Execute(os.Stdout, notes)
	if err != nil {
		log.Fatalf("Error executing template: %v", err)
	}
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
		Run:   listNotes(TemplateFormatter{}),
	}

	cmdList.Flags().StringVar(&templ, "template", "default", "The template to use for printing")

	cmdPromote := &cobra.Command{
		Use:   "promote",
		Short: "Promote a note",
		Run:   promoteNote,
	}

	cmdDemote := &cobra.Command{
		Use:   "demote",
		Short: "Demote a note",
		Run:   demoteNote,
	}

	cmdPriority := &cobra.Command{
		Use:   "priority <index> <priority>",
		Short: "Set note priority",
		Run:   updatePriority,
	}

	rootCmd.AddCommand(cmdAdd, cmdList, cmdRemove, cmdPromote, cmdDemote, cmdPriority)

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
		Priority: MEDIUM,
		State:    NONE,
		Contents: args[0],
	})

	slices.SortFunc(notes, func(a, b Note) int {
		return strings.Compare(a.Contents, b.Contents)
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
func listNotes(f Formatter) func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		path := cmd.Root().Annotations["stateFilePath"]
		notes := readState(path)

		f.format(os.Stdout, notes)
	}
}

// promoteNote promotes a note to the next state
func promoteNote(cmd *cobra.Command, args []string) {
	path := cmd.Root().Annotations["stateFilePath"]
	notes := readState(path)

	if len(args) < 1 {
		log.Fatal("usage: cmd-notes promote <index>")
	}

	index, err := strconv.Atoi(args[0])
	if err != nil {
		log.Fatal("error parsing note index: ", err)
	}

	if index >= len(notes) {
		log.Fatal("invalid note index")
	}

	switch notes[index].State {
	case NONE:
		notes[index].State = TODO
	case TODO:
		notes[index].State = IN_PROGRESS
	case IN_PROGRESS:
		notes[index].State = COMPLETE
	}

	writeState(path, &notes)

	fmt.Println("\033[32m \u2713 \033[0mpromoted note")
}

// demoteNote demotes a note to the previous state
func demoteNote(cmd *cobra.Command, args []string) {
	path := cmd.Root().Annotations["stateFilePath"]
	notes := readState(path)

	if len(args) < 1 {
		log.Fatal("usage: cmd-notes demote <index>")
	}

	index, err := strconv.Atoi(args[0])
	if err != nil {
		log.Fatal("error parsing note index: ", err)
	}

	if index >= len(notes) {
		log.Fatal("invalid note index")
	}

	switch notes[index].State {
	case TODO:
		notes[index].State = NONE
	case IN_PROGRESS:
		notes[index].State = TODO
	case COMPLETE:
		notes[index].State = IN_PROGRESS
	}

	writeState(path, &notes)

	fmt.Println("\033[32m \u2713 \033[0mdemoted note")
}

// updatePriority changes the priority of a note
func updatePriority(cmd *cobra.Command, args []string) {
	path := cmd.Root().Annotations["stateFilePath"]
	notes := readState(path)

	if len(args) < 2 {
		log.Fatal("usage: cmd-notes priority <index> <priority>")
	}

	index, err := strconv.Atoi(args[0])
	if err != nil {
		log.Fatal("error parsing note index: ", err)
	}

	if index >= len(notes) {
		log.Fatal("invalid note index")
	}

	priority, err := strconv.Atoi(args[1])
	if err != nil {
		log.Fatal("error parsing note priority: ", err)
	}

	if priority != LOW && priority != MEDIUM && priority != HIGH {
		log.Fatal("invalid note priority")
	}

	notes[index].Priority = priority

	writeState(path, &notes)

	fmt.Println("\033[32m \u2713 \033[0mpriority updated")
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
