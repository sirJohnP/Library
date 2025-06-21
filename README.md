# Library

Учебный проект с курса Go-2024. Реализация сервиса **library**.
Сервис поддерживает добавление и изменение информации о книгах и авторах.
Реализован паттерн outbox для взаимодействия со сторонними сервисами.

## Унификация технологий

Для удобства выполнения и проверки дз вводится ряд правил, унифицирующих используемые технологии

* Структура проекта [go-clean-template](https://github.com/evrone/go-clean-template) и
  этот [шаблон](https://github.com/itmo-org/lectures/tree/main/sem2/lecture1)
* Для генерации кода авторские [Makefile](./Makefile) и [easyp.yaml](./easyp.yaml)
* Для логирования [zap](https://github.com/uber-go/zap)
* Для валидации [protoc-gen-validate](https://github.com/bufbuild/protoc-gen-validate)
* Для поддержки REST-to-gRPC API [gRPC gateway](https://grpc-ecosystem.github.io/grpc-gateway/)
* Для миграций [goose](https://github.com/pressly/goose)
* [pgx](https://github.com/jackc/pgx) как драйвер для postgres

## Makefile

Для удобств локальной разработки был сделан [`Makefile`](Makefile). Имеются следующие команды:

Запустить полный цикл (линтер, тесты):

```bash 

make all

```

Запустить только тесты:

```bash

make test

``` 

Запустить линтер:

```bash

make lint

```

При разработке на Windows рекомендуется использовать [WSL](https://learn.microsoft.com/en-us/windows/wsl/install), чтобы

была возможность пользоваться вспомогательными скриптами.
