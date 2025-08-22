package commands

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Avdushin/secret/pkg/config"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// @ delete cmd
func DeleteKeyCmd() *cobra.Command {
	var force bool
	var noBackup bool

	cmd := &cobra.Command{
		Use:   "delete-key",
		Short: "Удаляет GPG-ключ проекта",
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.LoadConfig()

			// Если конфиг не загружается или ключ в конфиге пустой,
			// пытаемся найти ключ в GPG
			var keyID string
			if err != nil || cfg.GPGKey == "" {
				fmt.Println("ℹ️  В конфиге проекта не найден GPG-ключ")
				fmt.Println("🔍 Пытаемся найти ключ в GPG...")

				// Пытаемся автоматически определить ключ проекта
				autoKey, autoErr := autoDetectKey()
				if autoErr != nil {
					fmt.Println("❌ Не удалось найти GPG-ключ проекта")
					fmt.Println("Сначала выполните: secret init")
					os.Exit(1)
				}

				keyID = autoKey
				fmt.Printf("✅ Найден ключ в GPG: %s\n", keyID)
			} else {
				keyID = cfg.GPGKey
			}

			// Проверяем, существует ли ключ в GPG
			if !keyExistsInGPG(keyID) {
				fmt.Printf("❌ Ключ %s не найден в GPG\n", keyID)
				if cfg.GPGKey != "" {
					fmt.Println("Очищаем конфигурацию...")
					cfg.GPGKey = ""
					config.SaveConfig(cfg) // Игнорируем ошибку
				}
				fmt.Println("Выполните: secret init")
				os.Exit(1)
			}

			// Получаем информацию о ключе
			keyInfo, err := getKeyInfo(keyID)
			if err != nil {
				fmt.Printf("❌ Ошибка получения информации о ключе: %v\n", err)
				os.Exit(1)
			}

			if !force {
				fmt.Printf("\nВы собираетесь удалить ключ проекта:\n")
				fmt.Printf("ID: %s\n", keyID)
				fmt.Printf("Имя: %s\n", keyInfo.name)
				fmt.Printf("Email: %s\n", keyInfo.email)
				fmt.Print("\nПродолжить? (y/N): ")

				reader := bufio.NewReader(os.Stdin)
				confirm, _ := reader.ReadString('\n')
				if strings.ToLower(strings.TrimSpace(confirm)) != "y" {
					fmt.Println("Отмена удаления")
					return
				}
			}

			// Создаем резервную копию (если не отключено)
			if !noBackup {
				fmt.Println("\nСоздаем резервные копии ключей...")
				if err := createBackup(keyID); err != nil {
					fmt.Printf("⚠️ Не удалось создать резервную копию: %v\n", err)
					fmt.Println("Продолжаем без резервной копии")
				}
			}

			// Удаляем ключ из GPG
			fmt.Println("\nУдаляем ключ из GPG...")
			if err := deleteKey(keyID); err != nil {
				fmt.Printf("\n❌ Ошибка удаления ключа: %v\n", err)
				printManualDeleteInstructions(keyID)
				os.Exit(1)
			}

			// Удаляем ключ из конфига (если он там был)
			if cfg.GPGKey != "" {
				cfg.GPGKey = ""
				if err := config.SaveConfig(cfg); err != nil {
					fmt.Printf("⚠️ Ключ удален из GPG, но не удалось обновить конфиг: %v\n", err)
					os.Exit(1)
				}
			}

			fmt.Printf("\n✅ Ключ %s успешно удален из GPG\n", keyID)
			fmt.Println("Файлы секретов и резервные копии сохранены в директории .secrets/")
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Удалить без подтверждения")
	cmd.Flags().BoolVar(&noBackup, "no-backup", false, "Не создавать резервные копии ключей")
	return cmd
}

type keyInfo struct {
	name  string
	email string
}

// autoDetectKey пытается автоматически определить ключ проекта
func autoDetectKey() (string, error) {
	cmd := exec.Command("gpg", "--list-secret-keys", "--keyid-format", "LONG")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("gpg error: %v", err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "sec") && strings.Contains(line, "4096") {
			parts := strings.Fields(line)
			if len(parts) >= 4 {
				keyPart := parts[3]
				if strings.Contains(keyPart, "/") {
					keyParts := strings.Split(keyPart, "/")
					if len(keyParts) == 2 {
						return keyParts[1], nil
					}
				}
			}
		}
	}

	return "", fmt.Errorf("не удалось автоматически определить ключ")
}

// keyExistsInGPG проверяет, существует ли ключ в GPG
func keyExistsInGPG(keyID string) bool {
	cmd := exec.Command("gpg", "--list-keys", keyID)
	err := cmd.Run()
	return err == nil
}

func getKeyInfo(keyID string) (*keyInfo, error) {
	cmd := exec.Command("gpg", "--list-keys", "--with-colons", keyID)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("gpg error: %v", err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "uid:") {
			parts := strings.Split(line, ":")
			if len(parts) > 9 {
				return &keyInfo{
					name:  parts[9],
					email: extractEmail(parts[9]),
				}, nil
			}
		}
	}

	return nil, fmt.Errorf("не удалось разобрать информацию о ключе")
}

func extractEmail(uid string) string {
	start := strings.Index(uid, "<")
	end := strings.Index(uid, ">")
	if start >= 0 && end > start {
		return uid[start+1 : end]
	}
	return ""
}

func createBackup(keyID string) error {
	backupDir := filepath.Join(".secrets", "backup")
	if err := os.MkdirAll(backupDir, 0700); err != nil {
		return fmt.Errorf("не удалось создать директорию: %v", err)
	}

	// Экспорт публичного ключа
	pubKeyPath := filepath.Join(backupDir, fmt.Sprintf("key_%s.pub.asc", keyID))
	cmdPub := exec.Command("gpg", "--armor", "--export", keyID, "--output", pubKeyPath)
	if output, err := cmdPub.CombinedOutput(); err != nil {
		return fmt.Errorf("экспорт публичного ключа: %s: %v", string(output), err)
	}

	// Экспорт приватного ключа (с вводом пароля если нужно)
	privKeyPath := filepath.Join(backupDir, fmt.Sprintf("key_%s.priv.asc", keyID))
	cmdPriv := exec.Command("gpg", "--armor", "--export-secret-keys", keyID, "--output", privKeyPath)

	// Если терминал поддерживает ввод пароля
	if term.IsTerminal(int(os.Stdin.Fd())) {
		cmdPriv.Stdin = os.Stdin
		cmdPriv.Stdout = os.Stdout
		cmdPriv.Stderr = os.Stderr
	}

	if output, err := cmdPriv.CombinedOutput(); err != nil {
		return fmt.Errorf("экспорт приватного ключа: %s: %v", string(output), err)
	}

	fmt.Printf("✅ Резервные копии сохранены в %s\n", backupDir)
	return nil
}

func deleteKey(keyID string) error {
	// Удаляем приватный ключ
	cmdDelSecret := exec.Command("gpg", "--batch", "--yes", "--delete-secret-keys", keyID)
	cmdDelSecret.Stdin = os.Stdin
	cmdDelSecret.Stdout = os.Stdout
	cmdDelSecret.Stderr = os.Stderr

	if err := cmdDelSecret.Run(); err != nil {
		return fmt.Errorf("не удалось удалить приватный ключ: %v", err)
	}

	// Удаляем публичный ключ
	cmdDelPub := exec.Command("gpg", "--batch", "--yes", "--delete-keys", keyID)
	cmdDelPub.Stdin = os.Stdin
	cmdDelPub.Stdout = os.Stdout
	cmdDelPub.Stderr = os.Stderr

	if err := cmdDelPub.Run(); err != nil {
		return fmt.Errorf("не удалось удалить публичный ключ: %v", err)
	}

	return nil
}

func printManualDeleteInstructions(keyID string) {
	fmt.Println("\nПопробуйте выполнить следующие команды вручную:")
	fmt.Println()
	fmt.Printf("1. Удалить приватный ключ:\n   gpg --delete-secret-keys %s\n", keyID)
	fmt.Printf("2. Удалить публичный ключ:\n   gpg --delete-keys %s\n", keyID)
	fmt.Println()
	fmt.Println("Если возникают ошибки прав доступа, попробуйте с sudo:")
	fmt.Printf("   sudo gpg --delete-secret-keys %s\n", keyID)
	fmt.Printf("   sudo gpg --delete-keys %s\n", keyID)
	fmt.Println()
	fmt.Println("Если ключ защищен паролем, введите его при запросе")
}

// package commands

// import (
// 	"bufio"
// 	"fmt"
// 	"os"
// 	"os/exec"
// 	"path/filepath"
// 	"strings"

// 	"github.com/Avdushin/secret/pkg/config"
// 	"github.com/spf13/cobra"
// 	"golang.org/x/term"
// )

// // @ delete cmd
// func DeleteKeyCmd() *cobra.Command {
// 	var force bool
// 	var noBackup bool

// 	cmd := &cobra.Command{
// 		Use:   "delete-key",
// 		Short: "Удаляет GPG-ключ проекта",
// 		Run: func(cmd *cobra.Command, args []string) {
// 			cfg, err := config.LoadConfig()

// 			// Если конфиг не загружается или ключ в конфиге пустой,
// 			// пытаемся найти ключ в GPG
// 			var keyID string
// 			if err != nil || cfg.GPGKey == "" {
// 				fmt.Println("ℹ️  В конфиге проекта не найден GPG-ключ")
// 				fmt.Println("🔍 Пытаемся найти ключ в GPG...")

// 				// Пытаемся автоматически определить ключ проекта
// 				autoKey, autoErr := autoDetectKey()
// 				if autoErr != nil {
// 					fmt.Println("❌ Не удалось найти GPG-ключ проекта")
// 					fmt.Println("Сначала выполните: secret init")
// 					os.Exit(1)
// 				}

// 				keyID = autoKey
// 				fmt.Printf("✅ Найден ключ в GPG: %s\n", keyID)
// 			} else {
// 				keyID = cfg.GPGKey
// 			}

// 			// Проверяем, существует ли ключ в GPG
// 			if !keyExistsInGPG(keyID) {
// 				fmt.Printf("❌ Ключ %s не найден в GPG\n", keyID)
// 				if cfg.GPGKey != "" {
// 					fmt.Println("Очищаем конфигурацию...")
// 					cfg.GPGKey = ""
// 					config.SaveConfig(cfg) // Игнорируем ошибку
// 				}
// 				fmt.Println("Выполните: secret init")
// 				os.Exit(1)
// 			}

// 			// Получаем информацию о ключе
// 			keyInfo, err := getKeyInfo(keyID)
// 			if err != nil {
// 				fmt.Printf("❌ Ошибка получения информации о ключе: %v\n", err)
// 				os.Exit(1)
// 			}

// 			if !force {
// 				fmt.Printf("\nВы собираетесь удалить ключ проекта:\n")
// 				fmt.Printf("ID: %s\n", keyID)
// 				fmt.Printf("Имя: %s\n", keyInfo.name)
// 				fmt.Printf("Email: %s\n", keyInfo.email)
// 				fmt.Print("\nПродолжить? (y/N): ")

// 				reader := bufio.NewReader(os.Stdin)
// 				confirm, _ := reader.ReadString('\n')
// 				if strings.ToLower(strings.TrimSpace(confirm)) != "y" {
// 					fmt.Println("Отмена удаления")
// 					return
// 				}
// 			}

// 			// Создаем резервную копию (если не отключено)
// 			if !noBackup {
// 				fmt.Println("\nСоздаем резервные копии ключей...")
// 				if err := createBackup(keyID); err != nil {
// 					fmt.Printf("⚠️ Не удалось создать резервную копию: %v\n", err)
// 					fmt.Println("Продолжаем без резервной копии")
// 				}
// 			}

// 			// Удаляем ключ из GPG
// 			fmt.Println("\nУдаляем ключ из GPG...")
// 			if err := deleteKey(keyID); err != nil {
// 				fmt.Printf("\n❌ Ошибка удаления ключа: %v\n", err)
// 				printManualDeleteInstructions(keyID)
// 				os.Exit(1)
// 			}

// 			// Удаляем ключ из конфига (если он там был)
// 			if cfg.GPGKey != "" {
// 				cfg.GPGKey = ""
// 				if err := config.SaveConfig(cfg); err != nil {
// 					fmt.Printf("⚠️ Ключ удален из GPG, но не удалось обновить конфиг: %v\n", err)
// 					os.Exit(1)
// 				}
// 			}

// 			//! // Удаляем директорию с секретами
// 			// secretsDir := filepath.Join(".secrets")
// 			// if _, err := os.Stat(secretsDir); err == nil {
// 			// 	fmt.Println("🗑️  Удаляем директорию с секретами...")
// 			// 	os.RemoveAll(secretsDir)
// 			// }

// 			fmt.Printf("\n✅ Ключ %s успешно удален\n", keyID)
// 			fmt.Println("Проект полностью очищен от GPG-конфигурации")
// 		},
// 	}

// 	cmd.Flags().BoolVarP(&force, "force", "f", false, "Удалить без подтверждения")
// 	cmd.Flags().BoolVar(&noBackup, "no-backup", false, "Не создавать резервные копии ключей")
// 	return cmd
// }

// type keyInfo struct {
// 	name  string
// 	email string
// }

// // autoDetectKey пытается автоматически определить ключ проекта
// func autoDetectKey() (string, error) {
// 	cmd := exec.Command("gpg", "--list-secret-keys", "--keyid-format", "LONG")
// 	output, err := cmd.CombinedOutput()
// 	if err != nil {
// 		return "", fmt.Errorf("gpg error: %v", err)
// 	}

// 	lines := strings.Split(string(output), "\n")
// 	for _, line := range lines {
// 		if strings.Contains(line, "sec") && strings.Contains(line, "4096") {
// 			parts := strings.Fields(line)
// 			if len(parts) >= 4 {
// 				keyPart := parts[3]
// 				if strings.Contains(keyPart, "/") {
// 					keyParts := strings.Split(keyPart, "/")
// 					if len(keyParts) == 2 {
// 						return keyParts[1], nil
// 					}
// 				}
// 			}
// 		}
// 	}

// 	return "", fmt.Errorf("не удалось автоматически определить ключ")
// }

// // keyExistsInGPG проверяет, существует ли ключ в GPG
// func keyExistsInGPG(keyID string) bool {
// 	cmd := exec.Command("gpg", "--list-keys", keyID)
// 	err := cmd.Run()
// 	return err == nil
// }

// func getKeyInfo(keyID string) (*keyInfo, error) {
// 	cmd := exec.Command("gpg", "--list-keys", "--with-colons", keyID)
// 	output, err := cmd.CombinedOutput()
// 	if err != nil {
// 		return nil, fmt.Errorf("gpg error: %v", err)
// 	}

// 	lines := strings.Split(string(output), "\n")
// 	for _, line := range lines {
// 		if strings.HasPrefix(line, "uid:") {
// 			parts := strings.Split(line, ":")
// 			if len(parts) > 9 {
// 				return &keyInfo{
// 					name:  parts[9],
// 					email: extractEmail(parts[9]),
// 				}, nil
// 			}
// 		}
// 	}

// 	return nil, fmt.Errorf("не удалось разобрать информацию о ключе")
// }

// func extractEmail(uid string) string {
// 	start := strings.Index(uid, "<")
// 	end := strings.Index(uid, ">")
// 	if start >= 0 && end > start {
// 		return uid[start+1 : end]
// 	}
// 	return ""
// }

// func createBackup(keyID string) error {
// 	backupDir := filepath.Join(".secrets", "backup")
// 	if err := os.MkdirAll(backupDir, 0700); err != nil {
// 		return fmt.Errorf("не удалось создать директорию: %v", err)
// 	}

// 	// Экспорт публичного ключа
// 	pubKeyPath := filepath.Join(backupDir, fmt.Sprintf("key_%s.pub.asc", keyID))
// 	cmdPub := exec.Command("gpg", "--armor", "--export", keyID, "--output", pubKeyPath)
// 	if output, err := cmdPub.CombinedOutput(); err != nil {
// 		return fmt.Errorf("экспорт публичного ключа: %s: %v", string(output), err)
// 	}

// 	// Экспорт приватного ключа (с вводом пароля если нужно)
// 	privKeyPath := filepath.Join(backupDir, fmt.Sprintf("key_%s.priv.asc", keyID))
// 	cmdPriv := exec.Command("gpg", "--armor", "--export-secret-keys", keyID, "--output", privKeyPath)

// 	// Если терминал поддерживает ввод пароля
// 	if term.IsTerminal(int(os.Stdin.Fd())) {
// 		cmdPriv.Stdin = os.Stdin
// 		cmdPriv.Stdout = os.Stdout
// 		cmdPriv.Stderr = os.Stderr
// 	}

// 	if output, err := cmdPriv.CombinedOutput(); err != nil {
// 		return fmt.Errorf("экспорт приватного ключа: %s: %v", string(output), err)
// 	}

// 	fmt.Printf("✅ Резервные копии сохранены в %s\n", backupDir)
// 	return nil
// }

// func deleteKey(keyID string) error {
// 	// Удаляем приватный ключ
// 	cmdDelSecret := exec.Command("gpg", "--batch", "--yes", "--delete-secret-keys", keyID)
// 	cmdDelSecret.Stdin = os.Stdin
// 	cmdDelSecret.Stdout = os.Stdout
// 	cmdDelSecret.Stderr = os.Stderr

// 	if err := cmdDelSecret.Run(); err != nil {
// 		return fmt.Errorf("не удалось удалить приватный ключ: %v", err)
// 	}

// 	// Удаляем публичный ключ
// 	cmdDelPub := exec.Command("gpg", "--batch", "--yes", "--delete-keys", keyID)
// 	cmdDelPub.Stdin = os.Stdin
// 	cmdDelPub.Stdout = os.Stdout
// 	cmdDelPub.Stderr = os.Stderr

// 	if err := cmdDelPub.Run(); err != nil {
// 		return fmt.Errorf("не удалось удалить публичный ключ: %v", err)
// 	}

// 	return nil
// }

// func printManualDeleteInstructions(keyID string) {
// 	fmt.Println("\nПопробуйте выполнить следующие команды вручную:")
// 	fmt.Println()
// 	fmt.Printf("1. Удалить приватный ключ:\n   gpg --delete-secret-keys %s\n", keyID)
// 	fmt.Printf("2. Удалить публичный ключ:\n   gpg --delete-keys %s\n", keyID)
// 	fmt.Println()
// 	fmt.Println("Если возникают ошибки прав доступа, попробуйте с sudo:")
// 	fmt.Printf("   sudo gpg --delete-secret-keys %s\n", keyID)
// 	fmt.Printf("   sudo gpg --delete-keys %s\n", keyID)
// 	fmt.Println()
// 	fmt.Println("Если ключ защищен паролем, введите его при запросе")
// }
