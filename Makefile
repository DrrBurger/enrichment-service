.PHONY: build run clean test fmt vet lint help coverage-html coverage

## build: Билдит бинарный файл
build:
	go build -o bin/app -v cmd/enrich-app/main.go

## run_serv: Запускает сервер grpc
run:
	go run cmd/enrich-app/main.go

## clean: Очищяет и удаляет бинарный файл
clean:
	go clean
	rm -f bin/app

## fmt: Форматирование кода для соответствия стандартному стилю Go
fmt:
	go fmt ./...

## vet: Статический анализ кода на поиск подозрительных конструкций
vet:
	go vet ./...

## lint: Запускает линтер
lint:
	golangci-lint run

help: Makefile
	@echo " Choose a command run in "enrichment-service":"
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'