// This is an extemely simple package for creating and tracking
// short note snippets.
package main

import (
    "bufio"
    "fmt"
    "log"
    "os"
    "path/filepath"
    "runtime"
    "strconv"

    "github.com/spf13/cobra"
)

// Notes is a type alias for a simple string slice.
type Notes []string;

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
        Use: "add <note to add>",
        Short: "Add note",
        Run: addNote,
    }

    cmdRemove := &cobra.Command{
        Use: "rm <note index to remove>",
        Short: "Remove note",
        Run: removeNote,
    }

    cmdList := &cobra.Command{
        Use: "ls",
        Short: "List notes",
        Run: listNotes,
    }

    rootCmd.AddCommand(cmdAdd, cmdList, cmdRemove)

    return rootCmd
}

// addNote creates a new note.
func addNote(cmd *cobra.Command, args []string) {
    path := cmd.Root().Annotations["stateFilePath"]
    notes := readState(path)

    notes = append(notes, args[0])

    writeState(path, &notes)    

    fmt.Println("Successfully added note!")
}

// removeNote removes an existing note.
func removeNote(cmd *cobra.Command, args []string) {
    path := cmd.Root().Annotations["stateFilePath"]
    notes := readState(path)

    if len(args) < 1 {
        log.Fatal("specify index to remove")
    }

    index, err := strconv.Atoi(args[0])
    if err != nil {
        log.Fatal("error parsing note index: ", err)
    }

    if index >= len(notes) {
        log.Fatal("invalid note index")
    }

    notes = append(notes[:index], notes[index + 1:]...)

    writeState(path, &notes)
}

// listNotes lists all existing notes.
func listNotes(cmd *cobra.Command, args []string) {
    path := cmd.Root().Annotations["stateFilePath"]
    notes := readState(path)

    for index, note := range notes {
        fmt.Printf("%d) %s\n", index, note)
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
    notes := make([]string, 0)
    stateFilePath := filepath.Join(path, "state")

    file, err := os.Open(stateFilePath)
    if os.IsNotExist(err) {
        return notes
    } else if err != nil {
        log.Fatal("error opening state file: ", err) 
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)

    for scanner.Scan() {
        line := scanner.Text()
        notes = append(notes, line)
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
    if err != nil  {
        log.Fatal("error creating temporary state file: ", err) 
    }

    writer := bufio.NewWriter(file)

    for _, note := range *notes {
        fmt.Fprintln(writer, note)         
    }

    writer.Flush()

    file.Close()

    err = os.Rename(file.Name(), stateFilePath)
    if err != nil {
        log.Fatal("error renaming state file: ", err)
    }
}
