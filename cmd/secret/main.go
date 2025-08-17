package main

import (
	"fmt"
	"os"

	"github.com/Avdushin/secret/internal/commands"
	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:     "secret",
		Short:   "Утилита для управления секретами в проектах",
		Version: "0.1.0",
	}

	rootCmd.AddCommand(commands.InitCmd())
	rootCmd.AddCommand(commands.EncryptCmd())
	rootCmd.AddCommand(commands.DecryptCmd())
	rootCmd.AddCommand(commands.ExportKeyCmd())
	rootCmd.AddCommand(commands.DeleteKeyCmd())
	rootCmd.AddCommand(commands.CheckCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
