package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
    rootCmd := &cobra.Command{Use: "cmd-notes"}

    cmdAdd := &cobra.Command{
        Use: "add <note to add>",
        Short: "Add note",
        Run: func(cmd *cobra.Command, args []string) {
            fmt.Println("Adding a note")
        },
    }

    cmdRemove := &cobra.Command{
        Use: "rm <note id to remove>",
        Short: "Remove note",
        Run: func(cmd *cobra.Command, args []string) {
            fmt.Println("Removing note")
        },
    }

    cmdList := &cobra.Command{
        Use: "ls",
        Short: "List notes",
        Run: func(cmd *cobra.Command, args []string) {
            fmt.Println("Listing notes")
        },
    }

    rootCmd.AddCommand(cmdAdd, cmdList, cmdRemove)

    if err := rootCmd.Execute(); err != nil {
        fmt.Println(err)
        os.Exit(0)
    }
}
