# Makefile для проекта SteamProps

.PHONY: all build build-gui build-cli build-web test clean help

# Переменные
BINARY_NAME=steamprops
GUI_BINARY_NAME=steamprops-gui
CLI_BINARY_NAME=steamprops-cli
WEB_BINARY_NAME=steamprops-web

# Цель по умолчанию
all: build

# Сборка всех компонентов
build: build-cli build-gui build-web

# Сборка CLI приложения
build-cli:
	@echo "Сборка CLI приложения..."
	go build -o $(CLI_BINARY_NAME) cmd/main.go

# Сборка GUI приложения
build-gui:
	@echo "Сборка GUI приложения..."
	go build -o $(GUI_BINARY_NAME) cmd/gui/main.go cmd/gui/components.go cmd/gui/nogl.go

# Сборка веб-приложения
build-web:
	@echo "Сборка веб-приложения..."
	go build -o $(WEB_BINARY_NAME) cmd/web/main.go

# Запуск веб-приложения
run-web:
	@echo "Запуск веб-приложения..."
	./$(WEB_BINARY_NAME) 8080

# Запуск тестов
test:
	@echo "Запуск тестов..."
	go test ./...

# Запуск тестов с покрытием
test-coverage:
	@echo "Запуск тестов с покрытием..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Отчет о покрытии сохранен в coverage.html"

# Очистка
clean:
	@echo "Очистка..."
	rm -f $(BINARY_NAME) $(GUI_BINARY_NAME) $(CLI_BINARY_NAME) $(WEB_BINARY_NAME)
	rm -f coverage.out coverage.html

# Справка
help:
	@echo "Доступные команды:"
	@echo "  make build       - Сборка всех компонентов"
	@echo "  make build-cli   - Сборка CLI приложения"
	@echo "  make build-gui   - Сборка GUI приложения"
	@echo "  make build-web   - Сборка веб-приложения"
	@echo "  make run-web     - Запуск веб-приложения на порту 8080"
	@echo "  make test        - Запуск тестов"
	@echo "  make test-coverage - Запуск тестов с покрытием"
	@echo "  make clean       - Очистка собранных файлов"
	@echo "  make help        - Показать эту справку"
