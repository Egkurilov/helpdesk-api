# Helpdesk API

API для системы поддержки пользователей с функционалом создания тикетов, переписки и закрытия тикетов.

## Описание

Проект представляет собой RESTful API, разработанное на языке Go с использованием фреймворка [Gin](https://github.com/gin-gonic/gin) и ORM [GORM](https://gorm.io/). Оно предоставляет функционал для пользователей и операторов службы поддержки: создание тикетов, обмен сообщениями, просмотр истории переписки и закрытие тикетов. Авторизация реализована через JWT-токены, что обеспечивает безопасность доступа к защищенным маршрутам. База данных — PostgreSQL, а для удобного развертывания используется Docker и Docker Compose.

### Основные возможности
- **Пользователи**:
   - Регистрация и авторизация через Telegram ID.
   - Создание тикетов с указанием темы, описания и источника.
   - Отправка сообщений операторам по тикетам.
   - Закрытие собственных тикетов.
- **Операторы**:
   - Авторизация через логин и пароль.
   - Просмотр всех тикетов.
   - Отправка сообщений пользователям.
   - Закрытие любых тикетов.
- **Общее**:
   - Просмотр истории сообщений по тикету.
   - Логаут с подтверждением (для операторов).

## Требования

- Go 1.20 или выше
- Docker и Docker Compose
- PostgreSQL 13+ (для поддержки `gen_random_uuid()`)
- Git

## Установка

1. **Клонируй репозиторий:**
   ```bash
   git clone <repository_url>
   cd helpdesk-api