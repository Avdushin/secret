// internal/commands/init.go
package commands

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Avdushin/secret/pkg/config"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/gopenpgp/v2/helper"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// @ init cmd
func InitCmd() *cobra.Command {
	var backend string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Инициализирует проект для работы с секретами",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Инициализация с бэкендом: %s\n", backend)

			//@ имя текущей папки как имя проекта по умолчанию
			projectDir, err := os.Getwd()
			if err != nil {
				fmt.Printf("Ошибка получения пути: %v\n", err)
				os.Exit(1)
			}
			defaultProjectName := filepath.Base(projectDir)

			//? pr name
			projectName := promptUser(
				fmt.Sprintf("Название проекта [%s]: ", defaultProjectName),
				defaultProjectName,
			)

			//@ Запрашиваем файлы/директории для шифрования
			fmt.Println("\nУкажите файлы или директории для шифрования (через запятую)")
			fmt.Printf("По умолчанию: %s\n", strings.Join(config.DefaultSecretFiles, ", "))
			filesInput := promptUser("Файлы/директории: ", "")

			var secretFiles []string
			if filesInput == "" {
				secretFiles = config.DefaultSecretFiles
			} else {
				secretFiles = strings.Split(filesInput, ",")
				for i := range secretFiles {
					secretFiles[i] = strings.TrimSpace(secretFiles[i])
				}
			}

			//@ Запрашиваем параметры GPG ключа
			fmt.Println("\n⚙️  Настройка GPG ключа")

			// Выбор типа ключа
			keyType := promptUserWithOptions(
				"Тип ключа (RSA/ECC) [RSA]: ",
				[]string{"RSA", "ECC"},
				"RSA",
			)

			// Длина ключа
			var keyLength int
			var kt string
			switch keyType {
			case "RSA":
				keyLength = promptInt("Длина RSA ключа (2048/3072/4096) [4096]: ", 4096, []int{2048, 3072, 4096})
				kt = "rsa"
			case "ECC":
				keyLength = 0 // ECC использует кривые, а не длину
				kt = "x25519"
			}

			// Парольная фраза
			usePassphrase := promptYesNo("Использовать парольную фразу для ключа? (y/N): ", false)
			var passphrase string
			if usePassphrase {
				passphrase = promptPassword("Введите парольную фразу: ")
				confirm := promptPassword("Подтвердите парольную фразу: ")
				if passphrase != confirm {
					fmt.Println("❌ Парольные фразы не совпадают!")
					os.Exit(1)
				}
			}

			//@ Создаем GPG ключ
			keyName := fmt.Sprintf("%s Project Key", projectName)
			keyEmail := fmt.Sprintf("project+%s@team.org", strings.ToLower(projectName))

			fmt.Printf("\nСоздаем GPG-ключ для проекта: %s\n", keyName)
			privateKeyArmored, err := helper.GenerateKey(keyName, keyEmail, []byte(passphrase), kt, keyLength)
			if err != nil {
				fmt.Printf("Ошибка создания ключа: %v\n", err)
				os.Exit(1)
			}

			key, err := crypto.NewKeyFromArmored(privateKeyArmored)
			if err != nil {
				fmt.Printf("Ошибка создания ключа: %v\n", err)
				os.Exit(1)
			}

			publicKeyArmored, err := key.GetArmoredPublicKey()
			if err != nil {
				fmt.Printf("Ошибка получения публичного ключа: %v\n", err)
				os.Exit(1)
			}

			keyID := key.GetKeyID()

			// Save
			err = os.MkdirAll(".secret", 0700)
			if err != nil {
				fmt.Printf("Ошибка создания .secret: %v\n", err)
				os.Exit(1)
			}
			err = os.WriteFile(".secret/private.asc", []byte(privateKeyArmored), 0600)
			if err != nil {
				fmt.Printf("Ошибка сохранения приватного ключа: %v\n", err)
				os.Exit(1)
			}
			err = os.WriteFile(".secret/public.asc", []byte(publicKeyArmored), 0600)
			if err != nil {
				fmt.Printf("Ошибка сохранения публичного ключа: %v\n", err)
				os.Exit(1)
			}

			//@ Сохраняем конфиг
			cfg := &config.Config{
				Backend:     backend,
				GPGKey:      fmt.Sprintf("%X", keyID),
				ProjectName: projectName,
				SecretFiles: secretFiles,
			}

			if err := config.SaveConfig(cfg); err != nil {
				fmt.Printf("Ошибка сохранения конфига: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("\n✅ Успешно! Ключ создан (ID: %s)\n", keyID)
			fmt.Printf("🔑 Для экспорта ключа выполните: secret export\n")
			fmt.Printf("🔒 Для шифрования файлов выполните: secret encrypt\n")
		},
	}

	cmd.Flags().StringVarP(&backend, "backend", "b", "gpg", "Бэкенд (gpg, vault, bitwarden)")
	return cmd
}

func promptUser(prompt, defaultValue string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		return defaultValue
	}
	return input
}

func promptUserWithOptions(prompt string, options []string, defaultValue string) string {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print(prompt)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "" {
			return defaultValue
		}

		// Проверяем, есть ли ввод в допустимых опциях
		for _, option := range options {
			if strings.EqualFold(input, option) {
				return option
			}
		}

		fmt.Printf("❌ Неверный выбор. Допустимые варианты: %s\n", strings.Join(options, ", "))
	}
}

func promptInt(prompt string, defaultValue int, validValues []int) int {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print(prompt)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "" {
			return defaultValue
		}

		value, err := strconv.Atoi(input)
		if err != nil {
			fmt.Println("❌ Введите число")
			continue
		}

		// Если указаны допустимые значения, проверяем
		if len(validValues) > 0 {
			valid := false
			for _, v := range validValues {
				if value == v {
					valid = true
					break
				}
			}
			if !valid {
				fmt.Printf("❌ Неверное значение. Допустимые: %v\n", validValues)
				continue
			}
		}

		return value
	}
}

func promptYesNo(prompt string, defaultValue bool) bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print(prompt)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))

		if input == "" {
			return defaultValue
		}

		if input == "y" || input == "yes" || input == "д" || input == "да" {
			return true
		}
		if input == "n" || input == "no" || input == "н" || input == "нет" {
			return false
		}

		fmt.Println("❌ Пожалуйста, ответьте 'y' или 'n'")
	}
}

func promptPassword(prompt string) string {
	fmt.Print(prompt)

	// Пытаемся прочитать пароль без эха
	if term.IsTerminal(int(os.Stdin.Fd())) {
		bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err == nil {
			fmt.Println()
			return string(bytePassword)
		}
	}

	// Fallback: обычный ввод
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}
