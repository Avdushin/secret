### **Установка**  
```bash
# Linux/macOS
curl -L https://github.com/Avdushin/secret/releases/latest/download/secret-linux-amd64 -o /usr/local/bin/secret
chmod +x /usr/local/bin/secret
```

### **Пример использования**  
```bash
# 1. Инициализация
secret init --backend gpg

# 2. Шифрование
secret encrypt .env

# 3. Расшифровка
secret decrypt .env.gpg
```
