// internal/commands/check.go
package commands

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

// @ check cmd
func CheckCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "check",
		Short: "Проверяет доступность GPG ключей",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Проверяем доступные GPG ключи...")
			out, err := exec.Command("gpg", "--list-secret-keys", "--keyid-format=LONG").CombinedOutput()
			if err != nil {
				fmt.Printf("Ошибка: %v\n", err)
				return
			}
			fmt.Println(string(out))
		},
	}
}
