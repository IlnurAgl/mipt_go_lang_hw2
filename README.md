```
cd gateway
go run main.go
```

```
cd ledger
go run main.go
```

```
curl -X POST "127.0.0.1:8080/api/budget" -d '{"category": "test", "limit": 2000}'
{"Category":"test","Limit":2000}

curl -X GET "127.0.0.1:8080/api/budget"                                          
[{"Category":"test","Limit":2000}]

curl -X POST "127.0.0.1:8080/api/transaction" -d '{"amount": 2, "category": "test", "description": "test", "date": "2025-10-28"}'       
{"id":1,"amount":2,"category":"test","description":"test","date":"2025-10-28"}

 curl -X GET "127.0.0.1:8080/api/transaction"                                                                                            
[{"id":1,"amount":2,"category":"test","description":"","date":"2025-10-28"}]
```

```
ilnur@MacBook-Pro-Ilnur mipt_go_lang_hw2 %  curl -X POST "127.0.0.1:8080/api/reports/summary" -d '{"from": "2025-10-28", "to": "2025-10-29"}'
{"test":204}
ilnur@MacBook-Pro-Ilnur mipt_go_lang_hw2 %  curl -X POST "127.0.0.1:8080/api/reports/summary" -d '{"from": "2025-10-28", "to": "2025-10-29"}'
{"test":204
```

```
curl -X POST "127.0.0.1:8080/api/transactions/bulk" -d '
[
    {
        "amount": 1,
        "category": "еда",
        "description": "test",
        "date": "02-12-2025"
    },
    {
        "amount": 1,
        "category": "еда",
        "description": "test",
        "date": "02-12-2025"
    }
]'
```

## Миграции
```
goose -dir ./ledger/migrations postgres $DATABASE_URL up
goose -dir ./ledger/migrations postgres $DATABASE_URL down
```

## Переменные
```
DATABASE_URL
REDIS_ADDR
REDIS_PASSWORD
```
В postgres лежат данные о бюджетах и транзакциях, в redis кэщ отчетов с ttl 30s

## Тесты
```
go test ./... -cover -coverprofile=cover.out
go tool cover -html=cover.out -o cover.html
```
ledger 55.6%
gateway/main 54.5%