// internal/backends/gpg.go
package backends

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/Avdushin/secret/pkg/config"
	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/armor"
	"github.com/ProtonMail/go-crypto/openpgp/packet"
	"golang.org/x/term"
)

type GPGBackend struct {
	cfg *config.Config
}

func NewGPGBackend(cfg *config.Config) *GPGBackend {
	return &GPGBackend{cfg: cfg}
}

func (g *GPGBackend) Encrypt(file string) error {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return fmt.Errorf("файл %s не существует", file)
	}
	outFile := file + ".gpg"

	// Load public key
	pubArm, err := os.ReadFile(".secret/public.asc")
	if err != nil {
		return fmt.Errorf("не удалось прочитать публичный ключ: %v", err)
	}
	pubBlock, err := armor.Decode(bytes.NewReader(pubArm))
	if err != nil {
		return fmt.Errorf("ошибка декодирования публичного ключа: %v", err)
	}
	pubEntity, err := openpgp.ReadEntity(packet.NewReader(pubBlock.Body))
	if err != nil {
		return fmt.Errorf("ошибка чтения публичного ключа: %v", err)
	}

	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	out, err := os.Create(outFile)
	if err != nil {
		return err
	}
	defer out.Close()

	hints := &openpgp.FileHints{IsBinary: true, FileName: filepath.Base(file), ModTime: time.Now()}

	w, err := openpgp.Encrypt(out, []*openpgp.Entity{pubEntity}, nil, hints, nil)
	if err != nil {
		return fmt.Errorf("ошибка шифрования: %v", err)
	}

	_, err = io.Copy(w, f)
	if err != nil {
		return err
	}
	w.Close()

	// Создаем .example файл
	if err := createExampleFile(file); err != nil {
		return fmt.Errorf("не удалось создать .example файл: %v", err)
	}
	fmt.Printf("✅ Файл %s зашифрован в %s\n", file, outFile)
	return nil
}

func (g *GPGBackend) Decrypt(file string) error {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return fmt.Errorf("файл %s не существует", file)
	}
	outFile := strings.TrimSuffix(file, filepath.Ext(file))

	// Load private key
	privArm, err := os.ReadFile(".secret/private.asc")
	if err != nil {
		return fmt.Errorf("не удалось прочитать приватный ключ: %v", err)
	}
	privBlock, err := armor.Decode(bytes.NewReader(privArm))
	if err != nil {
		return fmt.Errorf("ошибка декодирования приватного ключа: %v", err)
	}
	privEntity, err := openpgp.ReadEntity(packet.NewReader(privBlock.Body))
	if err != nil {
		return fmt.Errorf("ошибка чтения приватного ключа: %v", err)
	}

	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	out, err := os.Create(outFile)
	if err != nil {
		return err
	}
	defer out.Close()

	prompt := func(keys []openpgp.Key, symmetric bool) ([]byte, error) {
		if symmetric {
			return nil, fmt.Errorf("симметричное шифрование не поддерживается")
		}
		// Предполагаем первый ключ
		pass := promptPassword("Введите парольную фразу для приватного ключа: ")
		err = keys[0].PrivateKey.Decrypt([]byte(pass))
		if err != nil {
			return nil, err
		}
		return nil, nil // nil passphrase, since asymmetric
	}

	md, err := openpgp.ReadMessage(f, openpgp.EntityList{privEntity}, prompt, nil)
	if err != nil {
		return fmt.Errorf("ошибка дешифровки: %v", err)
	}

	_, err = io.Copy(out, md.UnverifiedBody)
	if err != nil {
		return err
	}

	fmt.Printf("✅ Файл %s расшифрован в %s\n", file, outFile)
	return nil
}

func promptPassword(prompt string) string {
	fmt.Print(prompt)
	if term.IsTerminal(int(os.Stdin.Fd())) {
		bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err == nil {
			fmt.Println()
			return string(bytePassword)
		}
	}

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

// !TODO: вынести работу с .examples в отдельный модуль
func createExampleFile(originalFile string) error {
	content, err := os.ReadFile(originalFile)
	if err != nil {
		return err
	}
	// Определяем тип файла по расширению
	ext := filepath.Ext(originalFile)
	var processed string
	switch strings.ToLower(ext) {
	case ".env":
		processed = processEnvFile(string(content))
	case ".json":
		processed = processJSONFile(string(content))
	case ".yaml", ".yml":
		processed = processYAMLFile(string(content))
	case ".toml":
		processed = processTOMLFile(string(content))
	case ".ini":
		processed = processINIFile(string(content))
	default:
		// Для неизвестных форматов просто создаем пустой файл
		processed = "# Example file for " + filepath.Base(originalFile) + "\n"
	}
	// Формируем правильное имя example-файла
	dir := filepath.Dir(originalFile)
	fileBase := filepath.Base(originalFile)
	baseWithoutExt := strings.TrimSuffix(fileBase, ext)
	var exampleFileName string
	if strings.HasPrefix(fileBase, ".") {
		// Для скрытых файлов (.config.yaml) создаем .config.example.yaml
		parts := strings.SplitN(fileBase, ".", 3)
		if len(parts) >= 3 {
			exampleFileName = strings.Join(parts[:2], ".") + ".example." + strings.Join(parts[2:], ".")
		} else {
			exampleFileName = baseWithoutExt + ".example" + ext
		}
	} else {
		exampleFileName = baseWithoutExt + ".example" + ext
	}
	exampleFile := filepath.Join(dir, exampleFileName)
	return os.WriteFile(exampleFile, []byte(processed), 0644)
}

// Улучшенная обработка YAML файлов
func processYAMLFile(content string) string {
	// Обрабатываем простые и многострочные значения, заменяем на <placeholder> без кавычек
	re := regexp.MustCompile(`(?m)^(\s*[\w-]+\s*:\s*)(?:["'].*?['"]|\S+|>[^\n]*\n(?:\s+.*\n)*|\|[^\n]*\n(?:\s+.*\n)*)`)
	processed := re.ReplaceAllString(content, `${1}<placeholder>`)
	// Удаляем комментарии после значений
	processed = regexp.MustCompile(`(?m)^(\s*[\w-]+\s*:\s*<placeholder>)\s*#.*$`).ReplaceAllString(processed, `${1}`)
	return processed
}

func processEnvFile(content string) string {
	re := regexp.MustCompile(`(?m)^(\s*[\w-]+\s*=\s*)((?:"(.*?)")|(?:'(.*?)')|([^#\s]+))[ \t]*(.*)$`)
	processed := re.ReplaceAllStringFunc(content, func(m string) string {
		sub := re.FindStringSubmatch(m)
		if len(sub) < 7 {
			return m
		}
		prefix := sub[1]
		rest := sub[6]
		if sub[3] != "" {
			return prefix + "\"<placeholder>\"" + rest
		} else if sub[4] != "" {
			return prefix + "'<placeholder>'" + rest
		} else if sub[5] != "" {
			return prefix + "<placeholder>" + rest
		}
		return m
	})
	return processed
}

func processJSONFile(content string) string {
	return regexp.MustCompile(`(?m)"([\w-]+)"\s*:\s*"(?:[^"\\]|\\.)*"`).
		ReplaceAllString(content, `"${1}": "<placeholder>"`)
}

func processTOMLFile(content string) string {
	return regexp.MustCompile(`(?m)^\s*([\w-]+)\s*=\s*"(?:[^"\\]|\\.)*"`).
		ReplaceAllString(content, `${1} = "<placeholder>"`)
}

func processINIFile(content string) string {
	re := regexp.MustCompile(`(?m)^(\s*[\w-]+\s*=\s*)((?:"(.*?)")|(?:'(.*?)')|([^#\s]+))[ \t]*(.*)$`)
	processed := re.ReplaceAllStringFunc(content, func(m string) string {
		sub := re.FindStringSubmatch(m)
		if len(sub) < 7 {
			return m
		}
		prefix := sub[1]
		rest := sub[6]
		if sub[3] != "" {
			return prefix + "\"<placeholder>\"" + rest
		} else if sub[4] != "" {
			return prefix + "'<placeholder>'" + rest
		} else if sub[5] != "" {
			return prefix + "<placeholder>" + rest
		}
		return m
	})
	return processed
}
