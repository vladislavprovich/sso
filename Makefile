# Назва виконуваного файлу лінтера
LINTER = golangci-lint

# Мета за замовчуванням
all: lint

# Інсталяція golangci-lint'
install-linter:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Перевірка коду за допомогою лінтерів
lint:
	$(LINTER) run

# Автофікс проблем, які може виправити лінтер
lint-fix:
	$(LINTER) run --fix

# Чистка згенерованих файлів (опціонально)
clean:
	rm -rf $(LINTER)

# Допомога (виведе доступні команди)
help:
	@echo "Makefile для запуску Go лінтера"
	@echo "Доступні команди:"
	@echo "  install-linter  - інсталяція golangci-lint"
	@echo "  lint            - запуск лінтера для перевірки коду"
	@echo "  lint-fix        - запуск лінтера з автоматичним виправленням помилок"
	@echo "  clean           - чистка середовища (опціонально)"

.PHONY: all lint lint-fix clean help