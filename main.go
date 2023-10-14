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

    rootCmd.AddCommand(cmdAdd)

    if err := rootCmd.Execute(); err != nil {
        fmt.Println(err)
        os.Exit(0)
    }
}
