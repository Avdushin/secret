# Примеры использования утилиты Secret

[Типовой workflow работы с проектом](#типовой-workflow-работы-с-проектом)

Эта утилита предназначена для управления секретами в проектах с помощью GPG (и других бэкендов в будущем). Ниже приведены примеры использования каждой команды. Все примеры предполагают, что утилита установлена и вы находитесь в директории проекта.

## 1. Инициализация проекта
```bash
# Переходим в папку проекта
cd ~/projects/my-awesome-app

# Инициализируем secret (запрашивает название проекта и файлы для шифрования)
secret init

# Пример интерактивного ввода:
# Название проекта [my-awesome-app]: MyApp
# Файлы/директории: .env, config/*.yaml, secrets.json
```

## 2. Проверка ключей
```bash
# Проверяем ключ текущего проекта (по умолчанию)
secret check

# Показываем все доступные GPG ключи в системе
secret check --all

# Пример вывода:
# 🔍 Проверяем ключ проекта из конфига: ABCDEF1234567890
# ✅ Ключ проекта найден:
# sec   rsa4096/ABCDEF1234567890 2024-01-15 [SC]
# uid                 [ultimate] My Project (Project Key) <project@example.com>
# 🔐 Проверяем возможность шифрования... ✅ OK

# Пример когда проект не инициализирован, но ключи импортированы:
# 🔍 Автоматически определили ключ проекта: 0D141455C80A3F32
# ✅ Ключ проекта найден:
# sec   rsa4096/0D141455C80A3F32 2025-08-17 [SCEAR]
# uid               [ неизвестно ] prima-test Project Key (2025-08-17)
# 🔐 Проверяем возможность шифрования... ✅ OK
```

## 3. Шифрование файлов
```bash
# Шифруем конкретный файл
secret encrypt .env

# Шифруем все файлы, указанные при инициализации
secret encrypt

# Шифруем с явным указанием ключа
secret encrypt config/database.yaml -k OTHER_KEY_ID
```

## 4. Дешифровка файлов
```bash
# Дешифруем конкретный файл
secret decrypt .env.gpg

# Дешифруем все зашифрованные файлы проекта
secret decrypt

# Дешифруем в указанную директорию
secret decrypt prod.env.gpg > config/prod.env
```

## 5. Управление ключами
```bash
# Экспорт ключей проекта (в .secrets/backup/)
secret export

# Экспорт в конкретную директорию
secret export -o ~/backups/myapp-keys

# Импорт ключей проекта (автопоиск в текущей директории)
secret import

# Импорт из конкретной директории
secret import .secrets/backup/
secret import --dir ~/backups/myapp-keys

# Удаление ключа проекта (с подтверждением и бэкапом)
secret delete-key

# Принудительное удаление без подтверждения
secret delete-key --force
```

## 5.1 Импорт ключей

Чтобы импортировать GPG-ключи в свою систему:

```bash
# Автоматический поиск ключей в текущей директории
secret import

# Указать конкретную директорию для поиска
secret import --dir path/to/keys

# Принудительный импорт (если ключи уже существуют)
secret import --force
```

## 6. Работа с разными форматы
**Пример для .env:**
```bash
# Исходный файл:
# DB_PASSWORD="super-secret"
# API_KEY=123456

secret encrypt .env

# Создаст:
# .env.gpg - зашифрованная версия
# .env.example - с замененными значениями:
# DB_PASSWORD="<placeholder>"
# API_KEY="<placeholder>"
```

**Пример для config.yaml:**
```yaml
# Исходный файл:
database:
  host: "db.example.com"
  password: "qwerty" # секрет
```

`secret encrypt config.yaml`

```yaml
# Создаст:
# config.yaml.gpg - зашифрованная версия
# config.example.yaml - с замененными значениями:
database:
  host: "<placeholder>"
  password: "<placeholder>"
```

## 7. Интеграция с Git
```bash
# Добавляем в .gitignore чувствительные файлы
echo "
# Secret files
.secrets/
*.pub.asc
*.priv.asc
" >> .gitignore

# Коммитим и .gpg файлы и .example файлы
git add *.gpg *.example.*

# или просто git add .

git commit -m "Add encrypted configs and templates"

# Пушим всё включая зашифрованные файлы
git push
# Коммитим example-файлы
git add *.example.*
git commit -m "Add config templates"
```

## 8. Восстановление проекта
```bash
# Клонируем проект
git clone git@github.com:user/my-awesome-app.git
cd my-awesome-app

# Проверяем доступные ключи (опционально)
secret check --all

# Импортируем ключи проекта автоматически
secret import

# Проверяем, что ключ проекта загружен
secret check

# Дешифруем все файлы
secret decrypt
```

## 9. Продвинутые сценарии
**Шифрование нескольких файлов:**
```bash
# Шифруем все .env файлы
secret encrypt *.env

# Шифруем все файлы в config/
secret encrypt config/*
```

**Работа в CI/CD:**
```yaml
# .gitlab-ci.yml пример
deploy:
  before_script:
    - secret import
    - secret check  # Проверяем что ключ загружен
    - secret decrypt
  script:
    - ./deploy.sh
```

## 10. Получение справки
```bash
# Общая справка
secret --help

# Справка по команде
secret encrypt --help
secret check --help
```

## Особенности работы:
1. Для файлов с точкой в начале (например, `.config.yaml`) создаются корректные example-файлы (`.config.example.yaml`)
2. При дешифровке существующих файлов запрашивается подтверждение перезаписи
3. Поддерживаются сложные случаи в YAML:
   - Многострочные значения
   - Комментарии
   - Разные стили кавычек
4. Команда `check` помогает быстро проверить статус ключа проекта

Пример структуры проекта после работы:
```
my-awesome-app/
├── .secret/
│   └── config.yaml       # Конфиг утилиты
├── .env.example          # Шаблон
├── .env.gpg              # Зашифрованный файл
├── config/
│   ├── database.example.yaml
│   └── database.yaml.gpg
└── .secrets/
    └── backup/
        ├── myapp.priv.asc # Приватный ключ
        └── myapp.pub.asc  # Публичный ключ
```

## Типовой workflow работы с проектом

```bash
# 1. Клонируем проект
git clone project.git
cd project

# 2. Проверяем доступные ключи
secret check --all

# 3. Импортируем ключи проекта
secret import

# 4. Проверяем что ключ загружен
secret check

# 5. Дешифруем файлы
secret decrypt

# 6. Работаем с проектом...
# (вносим изменения в конфиги)

# 7. Шифруем файлы перед коммитом
secret encrypt

# 8. Коммитим только example-файлы
git add *.example.*
git commit -m "Update config templates"
```
