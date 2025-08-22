// internal/commands/check.go
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

// @ check cmd
func CheckCmd() *cobra.Command {
	var showAll bool

	cmd := &cobra.Command{
		Use:   "check",
		Short: "Проверяет доступность GPG ключей",
		Long: `Проверяет доступность GPG ключей.
По умолчанию показывает ключ текущего проекта.
С флагом --all показывает все доступные ключи.`,
		Run: func(cmd *cobra.Command, args []string) {
			if showAll {
				// Показываем все ключи
				checkAllKeys()
			} else {
				// Показываем ключ проекта
				checkProjectKey()
			}
		},
	}

	cmd.Flags().BoolVarP(&showAll, "all", "a", false, "Показать все доступные GPG ключи")
	return cmd
}

// ? все доступные GPG ключи
func checkAllKeys() {
	fmt.Println("🔍 Проверяем все доступные GPG ключи...")
	out, err := exec.Command("gpg", "--list-secret-keys", "--keyid-format=LONG").CombinedOutput()
	if err != nil {
		fmt.Printf("❌ Ошибка при получении списка ключей: %v\n", err)
		return
	}
	fmt.Println(string(out))
}

// ? Ключ текущего проекта
func checkProjectKey() {
	// Сначала пробуем загрузить конфиг проекта
	cfg, err := config.LoadConfig()
	var projectKey string
	if err == nil && cfg.GPGKey != "" {
		// Используем ключ из конфига
		projectKey = cfg.GPGKey
		fmt.Printf("🔍 Проверяем ключ проекта из конфига: %s\n", projectKey)
	} else {
		// Пытаемся автоматически определить ключ проекта по имени директории
		projectKey, err = detectProjectKeyFromDir()
		if err != nil {
			fmt.Printf("❌ Не удалось определить ключ проекта: %v\n", err)
			fmt.Println("Возможные решения:")
			fmt.Println("1. Выполните secret init для инициализации проекта")
			fmt.Println("2. Импортируйте ключи: secret import")
			fmt.Println("3. Укажите ключ вручную: secret check --all")
			os.Exit(1)
		}
		fmt.Printf("🔍 Автоматически определили ключ проекта: %s\n", projectKey)
	}

	// Проверяем существует ли ключ
	checkCmd := exec.Command("gpg", "--list-keys", projectKey)
	if output, err := checkCmd.CombinedOutput(); err != nil {
		fmt.Printf("❌ Ключ проекта не найден в GPG: %s\n", projectKey)
		fmt.Printf("Вывод: %s\n", string(output))
		fmt.Printf("Возможно ключ был удален или не импортирован\n")
		fmt.Println("Попробуйте импортировать ключ: secret import")
		os.Exit(1)
	} else {
		// Показываем информацию о ключе проекта
		fmt.Printf("✅ Ключ проекта найден:\n")
		// Получаем детальную информацию о ключе
		detailCmd := exec.Command("gpg", "--list-secret-keys", "--keyid-format=LONG", projectKey)
		if detailOutput, err := detailCmd.CombinedOutput(); err == nil {
			lines := strings.Split(string(detailOutput), "\n")
			for _, line := range lines {
				if strings.Contains(line, projectKey) || strings.Contains(line, "sec") || strings.Contains(line, "uid") {
					fmt.Println(line)
				}
			}
		}

		// Проверяем возможность шифрования/расшифровки
		fmt.Printf("\n🔐 Проверяем возможность шифрования... ")
		testEncryptCmd := exec.Command("gpg", "--encrypt", "--recipient", projectKey, "--armor", "--output", "/dev/null", "/dev/null")
		if err := testEncryptCmd.Run(); err != nil {
			fmt.Println("❌ Ошибка шифрования")
			fmt.Printf("Возможно ключ поврежден или не имеет необходимых прав\n")
		} else {
			fmt.Println("✅ OK")
		}
	}
}

// ? Пытаетмся определить ключ проекта по имени текущей директории
func detectProjectKeyFromDir() (string, error) {
	// Получаем имя текущей директории
	currentDir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dirName := filepath.Base(currentDir)

	// Получаем список всех ключей
	out, err := exec.Command("gpg", "--list-secret-keys", "--keyid-format=LONG").CombinedOutput()
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(out), "\n")
	for idx, line := range lines {
		if strings.Contains(line, "uid") && strings.Contains(line, dirName) {
			// Ищем "sec" в предыдущих строках (назад до 5 строк)
			for j := 1; j <= 5; j++ {
				if idx-j < 0 {
					break
				}
				prevLine := lines[idx-j]
				if strings.Contains(prevLine, "sec") {
					parts := strings.Fields(prevLine)
					if len(parts) >= 2 {
						keyPart := parts[1]
						if strings.Contains(keyPart, "/") {
							keyParts := strings.Split(keyPart, "/")
							if len(keyParts) == 2 {
								return keyParts[1], nil
							}
						}
					}
				}
			}
		}
	}
	return "", fmt.Errorf("не удалось найти ключ для проекта %s", dirName)
}
