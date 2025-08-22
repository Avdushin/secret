// internal/commands/export.go
package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Avdushin/secret/pkg/config"
	"github.com/spf13/cobra"
)

// @ export cmd
func ExportKeyCmd() *cobra.Command {
	var outputDir string

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Экспортирует GPG-ключ проекта",
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.LoadConfig()
			if err != nil {
				fmt.Printf("Ошибка загрузки конфига: %v\n", err)
				os.Exit(1)
			}

			if err := exportKeys(cfg, outputDir); err != nil {
				fmt.Printf("Ошибка экспорта ключей: %v\n", err)
				os.Exit(1)
			}
		},
	}

	cmd.Flags().StringVarP(&outputDir, "output", "o", "", "Директория для экспорта (по умолчанию .secrets/backup)")
	return cmd
}

func exportKeys(cfg *config.Config, outputDir string) error {
	if outputDir == "" {
		outputDir = filepath.Join(".secrets", "backup")
	}
	if err := os.MkdirAll(outputDir, 0700); err != nil {
		return fmt.Errorf("ошибка создания директории: %v", err)
	}

	// Формируем имя файла с именем проекта
	filenamePrefix := "key"
	if cfg.ProjectName != "" {
		filenamePrefix = strings.ToLower(strings.ReplaceAll(cfg.ProjectName, " ", "_"))
	}

	// Копируем публичный ключ
	pubKeyPath := filepath.Join(outputDir, fmt.Sprintf("%s.pub.asc", filenamePrefix))
	pubData, err := os.ReadFile(".secret/public.asc")
	if err != nil {
		return fmt.Errorf("не удалось прочитать публичный ключ: %v", err)
	}
	if err := os.WriteFile(pubKeyPath, pubData, 0600); err != nil {
		return fmt.Errorf("не удалось сохранить публичный ключ: %v", err)
	}

	// Копируем приватный ключ
	privKeyPath := filepath.Join(outputDir, fmt.Sprintf("%s.priv.asc", filenamePrefix))
	privData, err := os.ReadFile(".secret/private.asc")
	if err != nil {
		return fmt.Errorf("не удалось прочитать приватный ключ: %v", err)
	}
	if err := os.WriteFile(privKeyPath, privData, 0600); err != nil {
		return fmt.Errorf("не удалось сохранить приватный ключ: %v", err)
	}

	fmt.Printf("\n✅ Ключи экспортированы в %s:\n", outputDir)
	fmt.Printf(" - Публичный ключ: %s\n", pubKeyPath)
	fmt.Printf(" - Приватный ключ: %s\n", privKeyPath)
	fmt.Println("\n⚠️ Безопасно передайте приватный ключ другим участникам проекта!")

	return nil
}
