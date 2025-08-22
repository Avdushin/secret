# Примеры использования утилиты Secret

Эта утилита предназначена для управления секретами в проектах с помощью GPG (и других бэкендов в будущем). Ниже приведены примеры использования каждой команды. Все примеры предполагают, что утилита установлена и вы находитесь в директории проекта.

## 1. Инициализация проекта
```bash
# Переходим в папку проекта
cd ~/projects/my-awesome-app

# Инициализируем secret (запрашивает название проекта и файлы для шифрования)
./secret init

# Пример интерактивного ввода:
# Название проекта [my-awesome-app]: MyApp
# Файлы/директории: .env, config/*.yaml, secrets.json
```

## 2. Шифрование файлов
```bash
# Шифруем конкретный файл
./secret encrypt .env

# Шифруем все файлы, указанные при инициализации
./secret encrypt

# Шифруем с явным указанием ключа
./secret encrypt config/database.yaml -k OTHER_KEY_ID
```

## 3. Дешифровка файлов
```bash
# Дешифруем конкретный файл
./secret decrypt .env.gpg

# Дешифруем все зашифрованные файлы проекта
./secret decrypt

# Дешифруем в указанную директорию
./secret decrypt prod.env.gpg > config/prod.env
```

## 4. Управление ключами
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


## 4.1 Импорт ключей

Чтобы импортировать GPG-ключи в свою систему:

```bash
# Автоматический поиск ключей в текущей директории
secret import

# Указать конкретную директорию для поиска
secret import --dir path/to/keys

# Принудительный импорт (если ключи уже существуют)
secret import --force
```

## 5. Работа с разными форматами
**Пример для .env:**
```bash
# Исходный файл:
# DB_PASSWORD="super-secret"
# API_KEY=123456

./secret encrypt .env

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

./secret encrypt config.yaml

# Создаст:
# config.yaml.gpg - зашифрованная версия
# config.example.yaml - с замененными значениями:
database:
  host: "<placeholder>"
  password: "<placeholder>"
```

## 6. Интеграция с Git
```bash
# Добавляем в .gitignore
echo "
# Secret files
*.gpg
.secrets/
" >> .gitignore

# Коммитим example-файлы
git add *.example.*
git commit -m "Add config templates"
```

## 7. Восстановление проекта
```bash
# Клонируем проект
git clone git@github.com:user/my-awesome-app.git
cd my-awesome-app

# Импортируем ключ
gpg --import .secrets/backup/myapp.priv.asc

# Дешифруем все файлы
./secret decrypt
```

## 8. Продвинутые сценарии
**Шифрование нескольких файлов:**
```bash
# Шифруем все .env файлы
./secret encrypt *.env

# Шифруем все файлы в config/
./secret encrypt config/*
```

**Работа в CI/CD:**
```yaml
# .gitlab-ci.yml пример
deploy:
  before_script:
    - gpg --import ${GPG_PRIVATE_KEY}
    - ./secret decrypt
  script:
    - ./deploy.sh
```

## 9. Получение справки
```bash
# Общая справка
./secret --help

# Справка по команде
./secret encrypt --help
```

## Особенности работы:
1. Для файлов с точкой в начале (например, `.config.yaml`) создаются корректные example-файлы (`.config.example.yaml`)
2. При дешифровке существующих файлов запрашивается подтверждение перезаписи
3. Поддерживаются сложные случаи в YAML:
   - Многострочные значения
   - Комментарии
   - Разные стили кавычек

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