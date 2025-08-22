// internal/commands/check.go
package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"bytes"

	"github.com/Avdushin/secret/pkg/config"
	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/armor"
	"github.com/ProtonMail/go-crypto/openpgp/packet"
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
	// Since no system gpg, perhaps list from .secret
	priv, err := os.ReadFile(".secret/private.asc")
	if err != nil {
		fmt.Printf("❌ Нет ключей в .secret: %v\n", err)
		return
	}
	fmt.Println(string(priv))
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
	// Load private
	privArm, err := os.ReadFile(".secret/private.asc")
	if err != nil {
		fmt.Printf("❌ Ключ проекта не найден: %v\n", err)
		os.Exit(1)
	}

	block, err := armor.Decode(bytes.NewReader(privArm))
	if err != nil {
		fmt.Printf("❌ Ошибка декодирования: %v\n", err)
		os.Exit(1)
	}

	entity, err := openpgp.ReadEntity(packet.NewReader(block.Body))
	if err != nil {
		fmt.Printf("❌ Ошибка чтения ключа: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Ключ проекта найден:\n")
	fmt.Printf("KeyID: %s\n", entity.PrimaryKey.KeyIdString())
	fmt.Printf("Fingerprint: %s\n", entity.PrimaryKey.Fingerprint)
	for id, identity := range entity.Identities {
		fmt.Printf("Identity: %s\n", id, identity.Name)
	}
	fmt.Printf("Creation: %v\n", entity.PrimaryKey.CreationTime)

	// Проверяем возможность шифрования
	fmt.Printf("\n🔐 Проверяем возможность шифрования... ")

	// Test encrypt small data
	pubArm, err := os.ReadFile(".secret/public.asc")
	if err != nil {
		fmt.Println("❌ Ошибка")
		return
	}
	pubBlock, err := armor.Decode(bytes.NewReader(pubArm))
	if err != nil {
		fmt.Println("❌ Ошибка")
		return
	}
	pubEntity, err := openpgp.ReadEntity(packet.NewReader(pubBlock.Body))
	if err != nil {
		fmt.Println("❌ Ошибка")
		return
	}

	buf := bytes.NewBuffer(nil)
	w, err := openpgp.Encrypt(buf, []*openpgp.Entity{pubEntity}, nil, nil, nil)
	if err != nil {
		fmt.Println("❌ Ошибка шифрования")
		return
	}
	w.Write([]byte("test"))
	w.Close()

	fmt.Println("✅ OK")
}

func detectProjectKeyFromDir() (string, error) {
	// Получаем имя текущей директории
	currentDir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dirName := filepath.Base(currentDir)

	// Load private
	privArm, err := os.ReadFile(".secret/private.asc")
	if err != nil {
		return "", err
	}

	block, err := armor.Decode(bytes.NewReader(privArm))
	if err != nil {
		return "", err
	}

	entity, err := openpgp.ReadEntity(packet.NewReader(block.Body))
	if err != nil {
		return "", err
	}

	for _, identity := range entity.Identities {
		if strings.Contains(identity.Name, dirName) {
			return entity.PrimaryKey.KeyIdString(), nil
		}
	}
	return "", fmt.Errorf("не удалось найти ключ для проекта %s", dirName)
}
