# Отчет

Привет!
Если будут какие-то вопросы - смело пишите мне в тг @sirodgevalex

## **Что я сделал?**

1. Релизовал все предложенное api
2. По моим оценкам(последуют далее) условие №2 выполняется
3. Сделал пользователй и токены, регистрацию и авторизацию. Так как в самом Авито пользователей > 200 млн, то для них использовал бд (Постгрес). Активных пользователей(токены) храним в кэше (Редис). Токены выходят из строя через 5 минут. P S можно сделать, чтоб expired время у токенов обновлялось, но у нас не в этом суть задания. Редис я использовал, чтобы авторизация была быстрой (помним про 2 условие), а то в такой большой бд искаь каждый раз токен никакого времени не хватит
4. Релизовал Е2E-тест на сценарий получения баннера
5. Написал поход в бд напрямую в случае use_last_revision
6. Пока не сделал, в ближайшее время сделаю

## Пояснение про оптимизации

Чтобы пункт 2 был выполнен, пришлось оптимизировать некоторые части кода, а именно:

1) Кэш с активными пользователями для быстрой работы аутентификации (проверки токена на юзера или админа).
2) Кэш с баннерами, которые запрашивали (которые обновляли туда же). Если в этом кеше нет баннера, то придется идти в бд. По логике, в бд ходить будем не часто, поэтому пункт 2 выполнен.

## Интересности

1. Вернемся к авторизации,

## Как все запустить и что вообще делать?

0. Я прикрутил докер, поэтому можно делать просто

```bash
docker-compose up
```

Сами запросы я делал через curl, вот примеры

1. Запрос для создания пользователя (в апи не стал дописывать). Создаем юзера и админа

```bash
curl -i -X POST http://localhost:8082/register \
-H 'Content-Type: application/json' \
-d '{"email": "sirogdev@yandex.ru", "password": "Sneeejjir1_", "is_admin": false}'
```

```bash
curl -i -X POST http://localhost:8082/register \
-H 'Content-Type: application/json' \
-d '{"email": "kortkova@yandex.ru", "password": "REsdf12_", "is_admin": true}'
```

2. Запрос для аутентификации. В случае успеха вернется токен, он будет expired через 5 минут. Напоминаю про параллельную функцию удаления старых токенов из кэша.

```bash
curl -i -X POST http://localhost:8082/authorize \
-H 'Content-Type: application/json' \
-d '{"Email": "sirodgev@yandex.ru", "Password": "Sneeeir1_"}'
```

```bash
curl -i -X POST http://localhost:8082/authorize \
-H 'Content-Type: application/json' \
-d '{"email": "kortkova@yandex.ru", "password": "REsdf12_"}'
```

3. Запрос на создание баннера. Сюда токен надо вставить (тот, который вернулся в успешном ответе на авторизацию админа)

```bash
curl -i -X POST http://localhost:8082/banner \
-H 'Content-Type: application/json' \
-H 'token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJpc19hZG1pbiI6ZmFsc2UsImV4cCI6MTcxMzEyNzE5OSwic3ViIjoiMSJ9.uS_IMFpokzHGObZSZBMuJrPx_u8dWHNE_A3_YUcvrSg' \
-d '{
    "title": "Новый баннер1",
    "text": "Это новый баннер1",
    "url": "https://example.com/new_banner1",
    "feature_id": 1243,
    "tag_id": 457,
    "is_active": true
}'
```

4. Запрос для получения баннера для пользователя

```bash
curl -i -X GET http://localhost:8082/user_banner \
-H 'Content-Type: application/json' \
-H 'token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJpc19hZG1pbiI6ZmFsc2UsImV4cCI6MTcxMzA1ODc4OCwic3ViIjoiMSJ9.ImQeNyL7tCl28FyT0bKdE-0xIqA-n355vO1ReObpRU0' \
--data '{
  "tag_id": "457",
  "feature_id": "123",
  "use_last_revision": true
}'
```

9. Запрос для получения

```bash
curl -i -X GET http://localhost:8082/banner \
-H 'token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjozLCJpc19hZG1pbiI6dHJ1ZSwiZXhwIjoxNzEzMDg5MjAwLCJzdWIiOiIzIn0.EJHSImpvV9bc7JPFZYPN-HeTPmOoIpr50JpaMAK6dC0' \
--data '{
    "feature_id": 123
}'
```

10. Запрос для удаления

```bash
curl -i -X DELETE http://localhost:8082/banner/6 \
-H 'token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjozLCJpc19hZG1pbiI6dHJ1ZSwiZXhwIjoxNzEzMDg5MjAwLCJzdWIiOiIzIn0.EJHSImpvV9bc7JPFZYPN-HeTPmOoIpr50JpaMAK6dC0'
```

## Сложности, с которыми я столкунулся и как я их решил
