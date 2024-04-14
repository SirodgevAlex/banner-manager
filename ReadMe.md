добавить в апи создание пользователя, аутентификацию

1. Запрос для создания пользователя

```bash
curl -i -X POST http://localhost:8080/register \
-H 'Content-Type: application/json' \
-d '{"email": "sirogdev@yandex.ru", "password": "Sneeeir1_", "is_admin": false}'
```

```bash
curl -i -X POST http://localhost:8080/register \
-H 'Content-Type: application/json' \
-d '{"email": "kortkova@yandex.ru", "password": "REsdf12_", "is_admin": true}'
```

6. Запрос для аутентификации

```bash
curl -i -X POST http://localhost:8080/authorize \
-H 'Content-Type: application/json' \
-d '{"Email": "sirodgev@yandex.ru", "Password": "Sneeeir1_"}'
```

```bash
curl -i -X POST http://localhost:8080/authorize \
-H 'Content-Type: application/json' \
-d '{"email": "kortkova@yandex.ru", "password": "REsdf12_"}'
```

7. Запрос на создание баннера

```bash
curl -i -X POST http://localhost:8080/banner \
-H 'Content-Type: application/json' \
-H 'token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjozLCJpc19hZG1pbiI6dHJ1ZSwiZXhwIjoxNzEzMDY2Nzg4LCJzdWIiOiIzIn0.lZbJY1Wez1A0oANxH2Fm5AEx_HHP__Xw1pjnHgfvUGY' \
-d '{
    "title": "Новый баннер",
    "text": "Это новый баннер",
    "url": "https://example.com/new_banner",
    "feature_id": 123,
    "tag_id": 456,
    "is_active": true
}'
```

8. Запрос для получения баннера для пользователя

```bash
curl -i -X GET http://localhost:8080/user_banner \
-H 'Content-Type: application/json' \
-H 'token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJpc19hZG1pbiI6ZmFsc2UsImV4cCI6MTcxMzA1ODc4OCwic3ViIjoiMSJ9.ImQeNyL7tCl28FyT0bKdE-0xIqA-n355vO1ReObpRU0' \
--data '{
  "tag_id": "456",
  "feature_id": "123",
  "use_last_revision": true
}'
```

