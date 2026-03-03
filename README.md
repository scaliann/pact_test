# Telegram Multi-Session Service (Go + gotd + gRPC)

## Клонирование проекта

```bash
1. git init
2. git clone git@github.com:scaliann/pact_test.git
3. cd pact_test
```

## Настройка окржения

1. Создать `.env`, скопировав`.env.example`.

2. Заполнить `TELEGRAM_API_ID` and `TELEGRAM_API_HASH`.


## Запуск через Docker compose

```bash
docker compose up --build
```

## Как проверить:

- gRPC: `localhost:50051`
- UI: `http://localhost:8085`

Сквозная проверка через UI (покрывает требования задания):

1. Откройте `http://localhost:8085`.
2. В блоке **Session** нажмите `Create Session`.
3. Убедитесь, что в ответе есть `session_id` и `qr_code` (в блоке Session output).
4. Отсканируйте QR в мобильном Telegram: `Settings -> Devices -> Scan QR`.
5. Если QR истек, нажмите `Refresh QR` и отсканируйте новый код.
6. В блоке **SendMessage** заполните `Peer` (например `@durov` или username второго аккаунта) и `Text`, затем нажмите `Send Message`.
7. Убедитесь, что в ответе есть `message_id` (сообщение отправлено).
8. В блоке **Incoming Messages** нажмите `Start Subscribe`.
9. Отправьте сообщение на авторизованный аккаунт с другого Telegram-аккаунта и проверьте, что в UI появился новый элемент с полями `message_id`, `from`, `text`, `timestamp`.
10. Нажмите `Delete Session`.
11. Проверьте, что удаленная сессия действительно неактивна: `Send Message` / `Start Subscribe` со старым `session_id` должны возвращать ошибку.
12. Откройте второе окно браузера (или Incognito), создайте вторую сессию и авторизуйте другой Telegram-аккаунт.
13. Запустите `Subscribe` в обоих окнах.
14. Отправьте сообщения из обоих окон почти одновременно и проверьте, что обе отправки успешны.
15. Убедитесь, что каждое окно получает только свои обновления.
16. В одном окне отправьте сообщение на невалидный `Peer` и проверьте, что ошибка возникает только в этой сессии, а вторая продолжает нормально отправлять и получать сообщения.
17. Удалите одну сессию и проверьте, что вторая продолжает работать.




## Tests

```bash
go test ./...
```




