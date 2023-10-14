package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

type Notes []string;

func main() {
    command := setup()

    if err := command.Execute(); err != nil {
        log.Fatal(err)
    }
}

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

func addNote(cmd *cobra.Command, args []string) {
    path := cmd.Root().Annotations["stateFilePath"]
    notes := readState(path)

    notes = append(notes, args[0])

    writeState(path, &notes)    

    fmt.Println("Successfully added note!")
}

func removeNote(cmd *cobra.Command, args []string) {
    fmt.Println("Removing note")
}

func listNotes(cmd *cobra.Command, args []string) {
    path := cmd.Root().Annotations["stateFilePath"]
    notes := readState(path)

    for index, note := range notes {
        fmt.Printf("%d) %s\n", index, note)
    }
}

func getDataBasePath() string {
    basePath, exists := os.LookupEnv("XDG_STATE_HOME")

    if !exists {
        homeDir, err := os.UserHomeDir()
        if err != nil {
            log.Fatal("error locating user's home directory: ", err)
        }

        basePath = filepath.Join(homeDir, ".local/state")
    }

    basePath = filepath.Join(basePath, "cmd-notes")

    return basePath
}

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
