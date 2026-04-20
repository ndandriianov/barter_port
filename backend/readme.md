Конфигурация:

Порядок чтения и приоритет конфигов (от меньшего приоритета к большему):
1. Общий конфиг (`CommonPath`) - обязателен.
2. Сервисный конфиг (`ServicePath`) - опционален, мерджится поверх общего.
3. Конфиг окружения `./config/<APP_ENV>.yaml` - опционален, мерджится поверх общего и сервисного.
4. Переменные окружения - максимальный приоритет.

Итоговый приоритет: `общий -> сервисный -> env-файл -> переменные окружения`.

Сейчас через env явно биндятся ключи:
- `DB_PASSWORD`
- `JWT_ACCESS_SECRET`
- `JWT_REFRESH_SECRET`
- `KAFKA_USER_CREATION_TOPIC`
- `MAILER_BYPASS`
- `ADMIN_EMAIL`
- `ADMIN_PASSWORD`

Без явного биндинга эти переменные не будут учтены, даже если они присутствуют в окружении.
Это особенность библиотеки viper.


## Админ

При запуске `auth`-сервиса гарантируется наличие админа с дефолтными данными:
- Email: `admin@barterport.com`
- Пароль: `admin`

Значения берутся из `config/common.yaml`:
```yaml
admin:
  email: "admin@barterport.com"
  password: "admin"
```

При необходимости значения можно переопределить через `ADMIN_EMAIL` и `ADMIN_PASSWORD`.

Точка входа:
- `backend/cmd/auth/main.go`
- после загрузки конфига, миграций и инициализации приложения вызывается `authApp.EnsureAdmin(context.Background(), cfg.Admin.Email, cfg.Admin.Password)`

Где происходит логика:
- `backend/internal/auth/app/app.go`
- метод `EnsureAdmin(...)` делегирует вызов в `a.authService.CreateAdmin(...)`
- `backend/internal/auth/application/service.go`
- метод `CreateAdmin(...)` нормализует email, хэширует пароль, создаёт `domain.NewUser(...)`, выставляет `EmailVerified = true` и сохраняет пользователя через `s.users.Create(...)` внутри транзакции
- там же вызывается `s.createUser(...)`, чтобы создать связанное событие/запись для user creation

Поведение при повторном запуске:
- если пользователь с таким email уже существует, новый админ не создаётся
- сервис пишет `admin already exists` и продолжает запуск

Почта админа сразу помечается подтверждённой в таблице `users`.

## Demo seed

Для заполнения локального стенда тестовыми данными есть команда:

```bash
make seed-demo
```

Если нужно полностью очистить все базы через миграции и заново заполнить demo-данные:

```bash
make reseed-demo
```

Что она делает:
- регистрирует несколько demo-пользователей;
- обновляет профили;
- создаёт объявления, группу объявлений, черновик сделки;
- создаёт активную и завершённую сделки;
- добавляет отзывы и сообщения в чаты.

Demo-аккаунты:
- `alice.demo@barterport.local`
- `bob.demo@barterport.local`
- `clara.demo@barterport.local`
- `dan.demo@barterport.local`
- `eva.demo@barterport.local`
- `fedor.demo@barterport.local`

Пароль по умолчанию: `password123`.
Если задан `SEED_PASSWORD`, seed использует его для этих же аккаунтов.

Для работы модерационных сценариев (жалобы на офферы, разрешение провалов сделок) seed логинится как администратор.
По умолчанию используются кредs из `config/common.yaml` (`admin@barterport.com` / `admin`).
Переопределить можно через `SEED_ADMIN_EMAIL` и `SEED_ADMIN_PASSWORD`.

По умолчанию команда ходит в `http://localhost:80`, то есть ожидает поднятый app-контур через `caddy`.
Для локального сценария требуется `MAILER_BYPASS=true`, иначе обычная клиентская регистрация не сможет залогиниться без подтверждения почты.

Полезные переменные:
- `SEED_BASE_URL`
- `SEED_PASSWORD`
- `SEED_ADMIN_EMAIL`
- `SEED_ADMIN_PASSWORD`
- `SEED_TIMEOUT`
- `SEED_POLL_INTERVAL`
