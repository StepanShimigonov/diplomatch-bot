## Требования

* Go 1.25
* Docker (для контейнеризации).
* Токен API от MAX
* Переменные окружения: TOKEN (API токен)

## Запуск локально (без Docker)

- Установка
    ```bash
    git clone https://github.com/StepanShimigonov/diplomatch-bot    
    cd app
    go mod tidy
    go build -o diplomatch-bot /.
    ```

- Запуск
    ```bash
    export TOKEN="ваш_токен"
    ./diplomatch-bot
    ```

## Запуск в Docker

- Запуск

    ```bash
    docker build -t diplomatch-bot-max:latest .

    docker run -d --name diplomatch-bot \
    -e TOKEN="ваш_токен" \
    diplomatch-bot-max:latest
    ```

- Логи
    ```bash
    docker logs diplomatch-bot
    ```

- Остановка
    ```bash
    docker stop diplomatch-bot
    ```