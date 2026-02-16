# mylinter

## Конфигурация чувствительных слов

Анализатор может загружать чувствительные ключевые слова из YAML-файла конфигурации.

Использование через CLI:

```bash
go run ./cmd/mylinter -config ./example/mylinter.yml ./...
```

Формат конфигурации:

```yaml
sensitive_keywords:
  # необязательно: полностью заменяет список по умолчанию
  values:
    - password
    - token

  # необязательно: добавляет значения к активному списку
  add:
    - session_id

  # необязательно: удаляет значения из активного списка
  remove:
    - token
```

## Плагин для golangci-lint

Сборка плагина:

```bash
make build-plugin
```

Запуск с примером конфигурации:

```bash
golangci-lint run -c ./example/.golangci.yml
```
