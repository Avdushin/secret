package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Avdushin/secret/internal/backends"
	"github.com/Avdushin/secret/pkg/config"
	"github.com/spf13/cobra"
)

// @ decrypt cmd
func DecryptCmd() *cobra.Command {
	var allFiles bool

	cmd := &cobra.Command{
		Use:   "decrypt [file]",
		Short: "Расшифровывает файлы",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.LoadConfig()
			if err != nil {
				fmt.Printf("Ошибка загрузки конфига: %v\n", err)
				os.Exit(1)
			}

			gpg := backends.NewGPGBackend(cfg)

			// Если указан конкретный файл
			if len(args) == 1 {
				if err := gpg.Decrypt(args[0]); err != nil {
					fmt.Printf("❌ Ошибка: %v\n", err)
					os.Exit(1)
				}
				return
			}

			// Расшифровываем все зашифрованные файлы из конфига
			filesToDecrypt := getEncryptedFiles(cfg.SecretFiles)
			if len(filesToDecrypt) == 0 {
				fmt.Println("ℹ️ Не найдено файлов для расшифровки")
				return
			}

			fmt.Printf("🔓 Расшифровываем %d файлов...\n", len(filesToDecrypt))
			for _, file := range filesToDecrypt {
				if err := gpg.Decrypt(file); err != nil {
					fmt.Printf("⚠️ Ошибка при расшифровке %s: %v\n", file, err)
				}
			}

			fmt.Println("✅ Все файлы обработаны")
		},
	}

	cmd.Flags().BoolVarP(&allFiles, "all", "a", false, "Расшифровать все файлы из конфига")
	return cmd
}

func getEncryptedFiles(patterns []string) []string {
	var result []string
	for _, pattern := range patterns {
		matches, _ := filepath.Glob(pattern + ".gpg")
		result = append(result, matches...)
	}
	return result
}
