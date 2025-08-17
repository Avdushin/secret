package backends

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Avdushin/secret/pkg/config"
)

type GPGBackend struct {
	cfg *config.Config
}

func NewGPGBackend(cfg *config.Config) *GPGBackend {
	return &GPGBackend{cfg: cfg}
}

func (g *GPGBackend) Encrypt(file string) error {
	if g.cfg.GPGKey == "" {
		return fmt.Errorf("не настроен GPG-ключ проекта. Сначала выполните: secret init")
	}
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return fmt.Errorf("файл %s не существует", file)
	}
	outFile := file + ".gpg"
	cmd := exec.Command(
		"gpg",
		"--encrypt",
		"--recipient", g.cfg.GPGKey,
		"--trust-model", "always",
		"--output", outFile,
		file,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ошибка шифрования: %v", err)
	}
	// Создаем .example файл
	if err := createExampleFile(file); err != nil {
		return fmt.Errorf("не удалось создать .example файл: %v", err)
	}
	fmt.Printf("✅ Файл %s зашифрован в %s\n", file, outFile)
	return nil
}

func (g *GPGBackend) Decrypt(file string) error {
	if g.cfg.GPGKey == "" {
		return fmt.Errorf("не настроен GPG-ключ проекта")
	}
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return fmt.Errorf("файл %s не существует", file)
	}
	outFile := strings.TrimSuffix(file, filepath.Ext(file))
	cmd := exec.Command(
		"gpg",
		"--decrypt",
		"--output", outFile,
		file,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ошибка дешифровки: %v", err)
	}
	fmt.Printf("✅ Файл %s расшифрован в %s\n", file, outFile)
	return nil
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
	baseName := strings.TrimSuffix(originalFile, ext)
	if strings.HasPrefix(filepath.Base(originalFile), ".") {
		// Для скрытых файлов (.config.yaml) создаем .config.example.yaml
		parts := strings.SplitN(filepath.Base(originalFile), ".", 3)
		if len(parts) >= 3 {
			baseName = strings.Join(parts[:2], ".") + ".example." + strings.Join(parts[2:], ".")
		} else {
			baseName = baseName + ".example" + ext
		}
	} else {
		baseName = baseName + ".example" + ext
	}
	exampleFile := filepath.Join(filepath.Dir(originalFile), baseName)
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
