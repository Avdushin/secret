package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Avdushin/secret/pkg/config"
	"github.com/spf13/cobra"
)

// @ export cmd
func ExportKeyCmd() *cobra.Command {
	var outputDir string

	cmd := &cobra.Command{
		Use:   "export-key",
		Short: "Экспортирует GPG-ключ проекта",
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.LoadConfig()
			if err != nil {
				fmt.Printf("Ошибка загрузки конфига: %v\n", err)
				os.Exit(1)
			}

			if cfg.GPGKey == "" {
				fmt.Println("❌ В проекте не настроен GPG-ключ")
				fmt.Println("Сначала выполните: secret init")
				os.Exit(1)
			}

			// Создаем директорию для экспорта
			if outputDir == "" {
				outputDir = filepath.Join(".secrets", "backup")
			}
			if err := os.MkdirAll(outputDir, 0700); err != nil {
				fmt.Printf("Ошибка создания директории: %v\n", err)
				os.Exit(1)
			}

			// Формируем имя файла с именем проекта
			filenamePrefix := "key"
			if cfg.ProjectName != "" {
				filenamePrefix = strings.ToLower(strings.ReplaceAll(cfg.ProjectName, " ", "_"))
			}

			// Экспортируем публичный ключ
			pubKeyPath := filepath.Join(outputDir, fmt.Sprintf("%s.pub.asc", filenamePrefix))
			cmdPub := exec.Command("gpg", "--output", pubKeyPath, "--armor", "--export", cfg.GPGKey)
			if output, err := cmdPub.CombinedOutput(); err != nil {
				fmt.Printf("Ошибка экспорта публичного ключа: %s\n", output)
				os.Exit(1)
			}

			// Экспортируем приватный ключ
			privKeyPath := filepath.Join(outputDir, fmt.Sprintf("%s.priv.asc", filenamePrefix))
			cmdPriv := exec.Command("gpg", "--output", privKeyPath, "--armor", "--export-secret-keys", cfg.GPGKey)
			cmdPriv.Stdin = os.Stdin // Для ввода пароля если нужно
			if output, err := cmdPriv.CombinedOutput(); err != nil {
				fmt.Printf("Ошибка экспорта приватного ключа: %s\n", output)
				os.Exit(1)
			}

			fmt.Printf("\n✅ Ключи экспортированы в %s:\n", outputDir)
			fmt.Printf(" - Публичный ключ: %s\n", pubKeyPath)
			fmt.Printf(" - Приватный ключ: %s\n", privKeyPath)
			fmt.Println("\n⚠️ Безопасно передайте приватный ключ другим участникам проекта!")
		},
	}

	cmd.Flags().StringVarP(&outputDir, "output", "o", "", "Директория для экспорта (по умолчанию .secrets/backup)")
	return cmd
}
