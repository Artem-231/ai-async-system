# Asynchronous AI Image Generator

Асинхронная микросервисная система для генерации изображений с использованием нейросетей (HuggingFace/Pollinations). 
Проект демонстрирует реализацию паттернов отказоустойчивости, работу с брокерами сообщений и контейнеризацию приложений.

![Go](https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-316192?style=for-the-badge&logo=postgresql&logoColor=white)
![RabbitMQ](https://img.shields.io/badge/RabbitMQ-FF6600?style=for-the-badge&logo=rabbitmq&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-2496ED?style=for-the-badge&logo=docker&logoColor=white)
![Python](https://img.shields.io/badge/Python-3776AB?style=for-the-badge&logo=python&logoColor=white)

## Особенности реализации
- **Worker Pool / Connection Pool:** Реализовано ограничение конкурентных соединений к PostgreSQL (`SetMaxOpenConns`, `SetMaxIdleConns`) для защиты базы данных от исчерпания пула сокетов.
- **Graceful Shutdown:** Корректное завершение работы API-шлюза с использованием контекста (`context.WithTimeout`) и перехватом системных сигналов (SIGINT, SIGTERM) для предотвращения потери обрабатываемых запросов.
- **Idempotency & Polling:** Асинхронный опрос статусов со стороны клиента без блокировки основного потока выполнения.

## Архитектура
Система состоит из независимых компонентов, взаимодействующих через очередь сообщений:

- **API Gateway (Go):** Принимает HTTP-запросы от клиента или бота, проводит первичную валидацию, регистрирует задачу в БД и асинхронно ставит ее в очередь (Non-blocking I/O).
- **Telegram Bot (Go):** Клиентский интерфейс. Принимает команды пользователей и асинхронно опрашивает API Gateway о статусе задачи, отправляя готовый результат в чат.
- **RabbitMQ:** Брокер сообщений. Обеспечивает персистентность задач и гарантированную доставку (Durable queues), предотвращая потерю данных при падении или перезагрузке воркеров.
- **AI Worker (Python):** Фоновый сервис, потребляющий задачи из очереди (prefetch_count=1). Осуществляет взаимодействие с API генерации изображений, обновляет статусы задач в БД и сохраняет артефакты в локальную файловую систему.
- **PostgreSQL:** Реляционная база данных для персистентного хранения и отслеживания конечных автоматов задач (состояния: pending, done, error).

## Стек технологий
- **Backend:** Golang, Python 3
- **Infrastructure:** Docker, Docker Compose
- **Message Broker:** RabbitMQ
- **Database:** PostgreSQL
- **AI Integration:** Pollinations AI API

## Запуск проекта локально

1. Склонируйте репозиторий:
```bash
git clone [https://github.com/Artem-231/ai-async-system.git](https://github.com/Artem-231/ai-async-system.git)
cd ai-async-system
```

2. Создайте файл `.env` в корневой директории проекта для безопасного хранения конфигурации и укажите необходимые переменные:
```env
DB_USER=postgres
DB_PASSWORD=your_password
TG_TOKEN=your_telegram_bot_token
```

3. Убедитесь, что в корневой папке проекта существует директория `images` (она монтируется в контейнеры как volume для сохранения результатов):
```bash
mkdir -p images
```

4. Выполните сборку и запуск контейнеров в фоновом режиме:
```bash
docker-compose up --build -d
```

## Использование системы
Система предоставляет два способа взаимодействия: через интеграцию с Telegram и через REST API.

### Вариант 1: Telegram Bot
После старта контейнеров бот инициализируется автоматически. Отправьте команду в чат бота для создания задачи:
```text
/generate cyberpunk cat in neon city
```
Бот асинхронно дождется выполнения задачи и вернет сгенерированное изображение.

### Вариант 2: Использование REST API

**1. Создание задачи (POST /task)**
Пример отправки задачи на генерацию изображения:
```bash
curl -X POST http://localhost:8080/task \
  -H "Content-Type: application/json" \
  -d '{"payload": "cyberpunk cat in neon city"}'
```
В качестве ответа API вернет уникальный идентификатор задачи: `{"id": 1}`. Фоновый воркер асинхронно обработает запрос.

**2. Проверка статуса (GET /status)**
Для получения актуального статуса выполнения задачи используйте ранее полученный идентификатор:
```bash
curl -X GET "http://localhost:8080/status?id=1"
```
Ответ API содержит текущее состояние: `{"id": "1", "status": "done"}`. По завершении генерации итоговое изображение будет сохранено в директорию `images/1.png`.
