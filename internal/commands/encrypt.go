package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Avdushin/secret/internal/backends"
	"github.com/Avdushin/secret/pkg/config"
	"github.com/spf13/cobra"
)

// @ encrypt cmd
func EncryptCmd() *cobra.Command {
	var keyID string
	var allFiles bool

	cmd := &cobra.Command{
		Use:   "encrypt [file]",
		Short: "Шифрует конфигурационные файлы",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.LoadConfig()
			if err != nil {
				fmt.Printf("Ошибка загрузки конфига: %v\n", err)
				os.Exit(1)
			}

			// Если ключ указан явно, временно переопределяем
			if keyID != "" {
				cfg.GPGKey = keyID
			}

			gpg := backends.NewGPGBackend(cfg)

			// Если указан конкретный файл
			if len(args) == 1 {
				if err := gpg.Encrypt(args[0]); err != nil {
					fmt.Printf("❌ Ошибка: %v\n", err)
					os.Exit(1)
				}
				return
			}

			// Шифруем все файлы из конфига
			filesToEncrypt := getFilesToProcess(cfg.SecretFiles)
			if len(filesToEncrypt) == 0 {
				fmt.Println("ℹ️ Не найдено файлов для шифрования")
				return
			}

			fmt.Printf("🔒 Шифруем %d файлов...\n", len(filesToEncrypt))
			for _, file := range filesToEncrypt {
				if err := gpg.Encrypt(file); err != nil {
					fmt.Printf("⚠️ Ошибка при шифровании %s: %v\n", file, err)
				}
			}
			fmt.Println("✅ Все файлы обработаны")
		},
	}

	cmd.Flags().StringVarP(&keyID, "key", "k", "", "GPG Key ID для шифрования")
	cmd.Flags().BoolVarP(&allFiles, "all", "a", false, "Шифровать все файлы из конфига")
	return cmd
}

func getFilesToProcess(patterns []string) []string {
	var result []string
	for _, pattern := range patterns {
		matches, _ := filepath.Glob(pattern)
		result = append(result, matches...)
	}
	return result
}
