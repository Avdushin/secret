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

// ImportKeyCmd импортирует GPG-ключи проекта
func ImportKeyCmd() *cobra.Command {
	var keyDir string
	var force bool

	cmd := &cobra.Command{
		Use:   "import [directory]",
		Short: "Импортирует GPG-ключи проекта",
		Long: `Импортирует GPG-ключи проекта из указанной директории или автоматически
ищет ключи в текущей директории и поддиректориях.

Примеры:
  secret import                    # Автопоиск в текущей директории
  secret import .secrets/backup    # Поиск в указанной директории
  secret import --dir .secrets/backup # То же самое с флагом`,
		Args: cobra.MaximumNArgs(1), // Разрешаем 0 или 1 аргумент
		Run: func(cmd *cobra.Command, args []string) {
			// Обрабатываем аргумент командной строки (если передан)
			if len(args) > 0 && keyDir == "" {
				keyDir = args[0]
			}

			// Загружаем конфиг для получения имени проекта
			cfg, err := config.LoadConfig()
			if err != nil {
				fmt.Printf("Ошибка загрузки конфига: %v\n", err)
				os.Exit(1)
			}

			// Определяем префикс для поиска файлов ключей
			filenamePrefix := "key"
			if cfg.ProjectName != "" {
				filenamePrefix = strings.ToLower(strings.ReplaceAll(cfg.ProjectName, " ", "_"))
			}

			// Поиск файлов ключей
			pubKeyPath, privKeyPath, err := findKeyFiles(keyDir, filenamePrefix)
			if err != nil {
				fmt.Printf("❌ Ошибка поиска ключей: %v\n", err)
				fmt.Println("Попробуйте указать директорию с ключами: secret import <директория>")
				os.Exit(1)
			}

			fmt.Printf("🔍 Найдены ключи:\n")
			fmt.Printf(" - Публичный ключ: %s\n", pubKeyPath)
			fmt.Printf(" - Приватный ключ: %s\n", privKeyPath)

			// Импортируем публичный ключ
			fmt.Println("\n📥 Импортируем публичный ключ...")
			if err := importKey(pubKeyPath); err != nil {
				fmt.Printf("❌ Ошибка импорта публичного ключа: %v\n", err)
				os.Exit(1)
			}

			// Импортируем приватный ключ
			fmt.Println("📥 Импортируем приватный ключ...")
			if err := importKey(privKeyPath); err != nil {
				fmt.Printf("❌ Ошибка импорта приватного ключа: %v\n", err)
				os.Exit(1)
			}

			fmt.Println("\n✅ Ключи успешно импортированы!")
			fmt.Println("Теперь вы можете работать с зашифрованными файлами проекта.")
		},
	}

	cmd.Flags().StringVarP(&keyDir, "dir", "d", "", "Директория для поиска ключей (по умолчанию текущая директория)")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Принудительный импорт, даже если ключи уже существуют")

	return cmd
}

// findKeyFiles ищет файлы ключей в указанной директории
func findKeyFiles(searchDir, prefix string) (string, string, error) {
	if searchDir == "" {
		searchDir = "."
	}

	var pubKeyPath, privKeyPath string
	foundPub, foundPriv := false, false

	err := filepath.Walk(searchDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		filename := strings.ToLower(info.Name())

		// Ищем публичный ключ
		if !foundPub && strings.Contains(filename, prefix) &&
			(strings.HasSuffix(filename, ".pub.asc") ||
				strings.HasSuffix(filename, "_pub.asc") ||
				strings.Contains(filename, "public")) {
			pubKeyPath = path
			foundPub = true
		}

		// Ищем приватный ключ
		if !foundPriv && strings.Contains(filename, prefix) &&
			(strings.HasSuffix(filename, ".priv.asc") ||
				strings.HasSuffix(filename, "_priv.asc") ||
				strings.HasSuffix(filename, ".private.asc") ||
				strings.Contains(filename, "private")) {
			privKeyPath = path
			foundPriv = true
		}

		// Если нашли оба ключа, можно прервать поиск
		if foundPub && foundPriv {
			return filepath.SkipAll
		}

		return nil
	})

	if err != nil {
		return "", "", err
	}

	if !foundPub || !foundPriv {
		return "", "", fmt.Errorf("не найдены оба файла ключей. Искали файлы с префиксом: '%s' в директории: '%s'", prefix, searchDir)
	}

	return pubKeyPath, privKeyPath, nil
}

// importKey импортирует ключ с помощью GPG
func importKey(keyPath string) error {
	cmd := exec.Command("gpg", "--import", keyPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ошибка выполнения gpg --import: %v", err)
	}

	return nil
}
