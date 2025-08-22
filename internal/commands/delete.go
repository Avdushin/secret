// internal/commands/delete.go
package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"bytes"

	"github.com/Avdushin/secret/pkg/config"
	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/armor"
	"github.com/ProtonMail/go-crypto/openpgp/packet"
	"github.com/spf13/cobra"
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
			if err != nil {
				fmt.Printf("Ошибка загрузки конфига: %v\n", err)
				os.Exit(1)
			}

			// Load key info
			keyID, name, email, err := loadKeyInfo()
			if err != nil {
				fmt.Printf("❌ Ошибка получения информации о ключе: %v\n", err)
				os.Exit(1)
			}

			reader := bufio.NewReader(os.Stdin)

			if !force {
				fmt.Printf("\nВы собираетесь удалить ключ проекта:\n")
				fmt.Printf("ID: %s\n", keyID)
				fmt.Printf("Имя: %s\n", name)
				fmt.Printf("Email: %s\n", email)
				fmt.Print("\nПродолжить? (y/N): ")

				confirm, _ := reader.ReadString('\n')
				if strings.ToLower(strings.TrimSpace(confirm)) != "y" {
					fmt.Println("Отмена удаления")
					return
				}
			}

			// Обработка резервной копии
			if !noBackup {
				var doBackup bool
				if force {
					doBackup = true
				} else {
					fmt.Print("\nСделать резервную копию ключей перед удалением? (y/N): ")
					confirmBackup, _ := reader.ReadString('\n')
					doBackup = strings.ToLower(strings.TrimSpace(confirmBackup)) == "y"
				}

				if doBackup {
					fmt.Println("\nСоздаем резервные копии ключей...")
					if err := exportKeys(cfg, ""); err != nil {
						fmt.Printf("⚠️ Не удалось создать резервную копию: %v\n", err)
						fmt.Println("Продолжаем без резервной копии")
					}
				}
			}

			// Удаляем файлы ключей
			fmt.Println("\nУдаляем ключ из проекта...")
			err = os.Remove(".secret/private.asc")
			if err != nil {
				fmt.Printf("⚠️ Не удалось удалить приватный ключ: %v\n", err)
			}
			err = os.Remove(".secret/public.asc")
			if err != nil {
				fmt.Printf("⚠️ Не удалось удалить публичный ключ: %v\n", err)
			}

			// Удаляем ключ из конфига
			cfg.GPGKey = ""
			if err := config.SaveConfig(cfg); err != nil {
				fmt.Printf("⚠️ Не удалось обновить конфиг: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("\n✅ Ключ %s успешно удален\n", keyID)
			fmt.Println("Файлы секретов и резервные копии сохранены в директории .secrets/")
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Удалить без подтверждения")
	cmd.Flags().BoolVar(&noBackup, "no-backup", false, "Не создавать резервные копии ключей")
	return cmd
}

func loadKeyInfo() (keyID, name, email string, err error) {
	privArm, err := os.ReadFile(".secret/private.asc")
	if err != nil {
		return "", "", "", err
	}

	block, err := armor.Decode(bytes.NewReader(privArm))
	if err != nil {
		return "", "", "", err
	}

	entity, err := openpgp.ReadEntity(packet.NewReader(block.Body))
	if err != nil {
		return "", "", "", err
	}

	keyID = entity.PrimaryKey.KeyIdString()

	for _, identity := range entity.Identities {
		name = identity.Name
		// Parse email from name <email>
		start := strings.Index(name, "<")
		end := strings.Index(name, ">")
		if start >= 0 && end > start {
			email = name[start+1 : end]
		}
		break // Take first
	}

	return keyID, name, email, nil
}
