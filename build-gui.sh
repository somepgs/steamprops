#!/bin/bash

# Скрипт сборки GUI приложения SteamProps

echo "Сборка GUI приложения SteamProps..."

# Переходим в директорию проекта
cd "$(dirname "$0")"

# Собираем GUI приложение
go build -o steamprops-gui cmd/gui/main.go cmd/gui/components.go cmd/gui/nogl.go

if [ $? -eq 0 ]; then
    echo "✅ GUI приложение успешно собрано: steamprops-gui"
    echo "Для запуска выполните: ./steamprops-gui"
else
    echo "❌ Ошибка при сборке GUI приложения"
    exit 1
fi
