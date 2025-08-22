# Secret CLI

[![Go](https://img.shields.io/badge/Go-1.20%2B-blue?logo=go)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Version](https://img.shields.io/badge/Version-0.0.1-green)](https://github.com/Avdushin/secret/releases)

Командная утилита для управления секретами в проектах. Шифрует конфиденциальные файлы (`.env`, JSON, YAML и др.) с помощью GPG (планируется Vault/Bitwarden). Создаёт шаблоны `.example` с плейсхолдерами для безопасной совместной работы.

## :tada: Мотивация

Хранение секретов (API-ключи, пароли) в открытом виде рискованно. **Secret** шифрует файлы, генерирует примеры и упрощает обмен ключами, предотвращая утечки в Git без лишней сложности.

## :sparkles: Возможности

- :key: **Инициализация**: Генерация GPG-ключа и настройка секретных файлов.
- :lock: **Шифрование/Расшифровка**: Поддержка `.env`, JSON, YAML, TOML, INI; пакетные операции.
- :memo: **Шаблоны**: Авто-создание `.example` с `<placeholder>`.
- :key: **Ключи**: Экспорт, импорт, удаление, проверка.
- :globe_with_meridians: **Кросс-платформа**: Linux, macOS, Windows.
- :construction: **Расширяемость**: Бэкенды (GPG сейчас, Vault/Bitwarden скоро).

[Установка](#package-установка)

## :warning: Зависимости

Для работы программы требуется установленная утилита **GPG (GNU Privacy Guard)**:

### Linux
```bash
# Ubuntu/Debian
sudo apt update && sudo apt install gnupg

# Fedora/RHEL/CentOS
sudo dnf install gnupg2

# Arch Linux/Manjaro
sudo pacman -S gnupg

# Solus
sudo eopkg install gnupg

# Void Linux
sudo xbps-install gnupg

# Gentoo
sudo emerge --ask app-crypt/gnupg

# NixOS
nix-env -i gnupg
```

### macOS
```bash
# Homebrew
brew install gnupg

# MacPorts
sudo port install gnupg2
```

### Windows
- Скачайте [Gpg4win](https://www.gpg4win.org/)
- Или используйте [Chocolatey](https://chocolatey.org/):
  ```powershell
  choco install gpg4win
  ```
- Или [Scoop](https://scoop.sh/):
  ```powershell
  scoop install gpg
  ```

## :package: Установка

### Одной командой (Linux/macOS)

```bash
curl -sSL https://raw.githubusercontent.com/Avdushin/secret/main/install.sh | bash
```

### Ручная

1. Скачайте релиз с [GitHub](https://github.com/Avdushin/secret/releases) (e.g., `secret-0.0.1-linux-amd64`).
2. `chmod +x <file> && mv <file> /usr/local/bin/secret`.

Для Windows: Скачайте `.exe` и добавьте в PATH.

### Из исходников

```bash
git clone https://github.com/Avdushin/secret.git
cd secret
make build
```

## :books: Примеры использования

| Команда | Описание |
|---------|----------|
| `secret init` | Инициализация: создаёт ключ и конфиг. |
| `secret encrypt` | Шифрует все файлы, создаёт `.gpg` и `.example`. |
| `secret decrypt <file.gpg>` | Расшифровка файла. |
| `secret check` | Проверяет ключ проекта. |
| `secret check --all` | Показывает все доступные GPG ключи. |
| `secret export -o dir` | Экспорт ключей. |
| `secret import <dir>` | Импорт ключей. |
| `secret version` | Показ версии. |

Подробности в [docs/examples.md](docs/examples.md).

## :key: Управление ключами

```bash
# Экспорт ключей проекта (в .secrets/backup/)
./secret export

# Экспорт в конкретную директорию
./secret export -o ~/backups/myapp-keys

# Импорт ключей проекта (автопоиск в текущей директории)
./secret import

# Импорт из конкретной директории
./secret import .secrets/backup/
./secret import --dir ~/backups/myapp-keys

# Удаление ключа проекта (с подтверждением и бэкапом)
./secret delete-key

# Принудительное удаление без подтверждения
./secret delete-key --force
```

## :gear: Конфигурация

В `.secret/config.yaml`:

```yaml
backend: gpg
gpg_key: <ID>
secret_files:
  - .env
  - config.json
```

Редактируйте для кастомизации.

## :wrench: Разработка

- **Требования**: Go 1.20+.
- **Makefile**: `make build` (сборка), `make test` (тесты), `make release` (все платформы).
- **Структура**: Команды — `internal/commands/`, бэкенды — `internal/backends/`.

Детали в [docs/dev/Makefile.md](docs/dev/Makefile.md).

<p align="center">
    2025 © <a href="https://github.com/Avdushin" target="_blank">AVDUSHIN</a>
</p>
