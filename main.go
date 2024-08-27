// This is an extemely simple package for creating and tracking
// short note snippets.
package main

import (
	"cmd-notes/note"
	"cmd-notes/tui"
	"cmd-notes/utils"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"text/template"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var funcMap = template.FuncMap{
	"colorizeContents": func(state int, text string) string {
		var result string

		switch state {
		case note.TODO:
			result = note.RED + text + note.RESET
		case note.IN_PROGRESS:
			result = note.YELLOW + text + note.RESET
		case note.REVIEWING:
			result = note.BLUE + text + note.RESET
		case note.COMPLETE:
			result = note.GREEN + text + note.RESET
		default:
			result = text
		}

		return result
	},
	"priority": func(priority int) string {
		var result string

		switch priority {
		case note.LOW:
			result = "(" + note.BLUE + "\u2193" + note.RESET + ")"
		case note.MEDIUM:
			result = "(-)"
		case note.HIGH:
			result = "(" + note.RED + "\u2191" + note.RESET + ")"
		default:
			result = ""
		}

		return result
	},
	"status": func(status int) string {
		var result string

		switch status {
		case note.TODO:
			result = "Todo"
		case note.IN_PROGRESS:
			result = "In Progress"
		case note.REVIEWING:
			result = "Reviewing"
		case note.COMPLETE:
			result = "Complete"
		default:
			result = ""
		}
		return result
	},
	"index": func(index int) string {
		return note.BLUE + strconv.Itoa(index) + note.RESET
	},
	"filterByPriority": func(notes []note.Note, priority int) []note.Note {
		var filteredItems []note.Note

		for _, note := range notes {
			if note.Priority == priority {
				filteredItems = append(filteredItems, note)
			}
		}

		return filteredItems
	},
	"monday": func(priority int) string {
		now := time.Now()

		offset := int(time.Monday - now.Weekday())
		if offset > 0 {
			offset -= 6
		}

		offset += (note.HIGH - priority) * 7

		lastMonday := now.AddDate(0, 0, offset)

		return lastMonday.Format("01/02/2006")
	},
}

var templ string

type Formatter interface {
	format(w io.Writer, notes []note.Note)
}

type TerminalFormatter struct{}

func (f TerminalFormatter) format(w io.Writer, notes []note.Note) {
	for index, n := range notes {
		color := "\033[0m"
		priority := ""

		switch n.Priority {
		case note.LOW:
			priority = "(\033[34m\u2193\033[0m)"
		case note.MEDIUM:
			priority = "(-)"
		case note.HIGH:
			priority = "(\033[31m\u2191\033[0m)"
		}

		switch n.State {
		case note.TODO:
			color = "\033[91m"
		case note.IN_PROGRESS:
			color = "\033[93m"
		case note.COMPLETE:
			color = "\033[92m"
		}

		fmt.Fprintf(w, "%s\033[94m %d \033[0m- %s%s\033[0m\n", priority, index, color, n.Contents)
	}
}

type TemplateFormatter struct{}

func (f TemplateFormatter) format(w io.Writer, notes []note.Note) {
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

	cmdTui := &cobra.Command{
		Use:   "tui",
		Short: "Run tui",
		Run:   startTui,
	}

	rootCmd.AddCommand(cmdAdd, cmdList, cmdRemove, cmdPromote, cmdDemote, cmdPriority, cmdTui)

	return rootCmd
}

// addNote creates a new note.
func addNote(cmd *cobra.Command, args []string) {
	path := cmd.Root().Annotations["stateFilePath"]
	notes := utils.ReadState(path)

	if len(args) < 1 {
		log.Fatal("usage: cmd-notes add \"<note>\"")
	}

	notes = append(notes, note.Note{
		Priority: note.MEDIUM,
		State:    note.NONE,
		Contents: args[0],
	})

	utils.WriteState(path, &notes)

	fmt.Println("\033[32m \u2713 \033[0madded note")
}

// removeNote removes an existing note.
func removeNote(cmd *cobra.Command, args []string) {
	path := cmd.Root().Annotations["stateFilePath"]
	notes := utils.ReadState(path)

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

	utils.WriteState(path, &notes)

	fmt.Println("\033[31m \u2717 \033[0mremoved note")
}

// listNotes lists all existing notes.
func listNotes(f Formatter) func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		path := cmd.Root().Annotations["stateFilePath"]
		notes := utils.ReadState(path)

		f.format(os.Stdout, notes)
	}
}

// promoteNote promotes a note to the next state
func promoteNote(cmd *cobra.Command, args []string) {
	path := cmd.Root().Annotations["stateFilePath"]
	notes := utils.ReadState(path)

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

	notes[index].Promote()

	utils.WriteState(path, &notes)

	fmt.Println("\033[32m \u2713 \033[0mpromoted note")
}

// demoteNote demotes a note to the previous state
func demoteNote(cmd *cobra.Command, args []string) {
	path := cmd.Root().Annotations["stateFilePath"]
	notes := utils.ReadState(path)

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

	notes[index].Demote()

	utils.WriteState(path, &notes)

	fmt.Println("\033[32m \u2713 \033[0mdemoted note")
}

// updatePriority changes the priority of a note
func updatePriority(cmd *cobra.Command, args []string) {
	path := cmd.Root().Annotations["stateFilePath"]
	notes := utils.ReadState(path)

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

	if priority != note.LOW && priority != note.MEDIUM && priority != note.HIGH {
		log.Fatal("invalid note priority")
	}

	notes[index].Priority = priority

	utils.WriteState(path, &notes)

	fmt.Println("\033[32m \u2713 \033[0mpriority updated")
}

func startTui(cmd *cobra.Command, args []string) {
	path := cmd.Root().Annotations["stateFilePath"]
	notes := utils.ReadState(path)

	p := tea.NewProgram(tui.InitModel(path, notes), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("error: %v", err)
		os.Exit(1)
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
