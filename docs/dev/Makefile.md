### Как использовать:

1. **Базовая сборка** (для текущей платформы):
```bash
make build
# Результат: bin/secret
```

2. **Сборка для всех платформ**:
```bash
make release
# Результат:
# bin/secret-1.0.0-linux-amd64
# bin/secret-1.0.0-linux-arm64
# bin/secret-1.0.0-windows-amd64.exe
# bin/secret-1.0.0-windows-arm64.exe
# bin/secret-1.0.0-darwin-amd64
# bin/secret-1.0.0-darwin-arm64
```

3. **Сборка для конкретной ОС**:
```bash
make linux    # Только Linux версии
make windows  # Только Windows версии
make darwin   # Только macOS версии
```

4. **Другие полезные команды**:
```bash
make clean    # Очистить папку bin
make test     # Запустить тесты
make install  # Установить в систему
make lint     # Проверить код
```

### Особенности:
- Автоматически определяет версию из git тегов
- Оптимизированные бинарники (с флагами `-s -w`)
- Поддержка как amd64, так и arm64 архитектур
- Для Windows добавляется расширение .exe
- Отдельные цели для каждой ОС
- Интеграция с golangci-lint для проверки кода
