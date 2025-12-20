# Запуск
```
docker-compose up -d
```
Для auth + ledger дополнительно в их директориях 
```
goose -dir ./migrations postgres $DATABASE_URL up
```

# Endpoints
## Auth
```
grpcurl -plaintext -d '{"login": "testuser", "password": "testpass"}' \
  localhost:50052 auth.v1.AuthService.Register
{}

grpcurl -plaintext -d '{"login": "testuser", "password": "testpass"}' \
  localhost:50052 auth.v1.AuthService.Login 
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiOThhYzJmZmUtNTc0OS00YWRlLTkwNTMtNmE3NTU1ZDJiZjA0IiwiZXhwIjoxNzY2MzQ1NTkwLCJpYXQiOjE3NjYyNTkxOTB9.HuEGJC4maaSVSNZwwFOksYHKo6B-THgAQoH-S8fItQk"
}

grpcurl -plaintext -d '{"token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiOThhYzJmZmUtNTc0OS00YWRlLTkwNTMtNmE3NTU1ZDJiZjA0IiwiZXhwIjoxNzY2MzQ1NTkwLCJpYXQiOjE3NjYyNTkxOTB9.HuEGJC4maaSVSNZwwFOksYHKo6B-THgAQoH-S8fItQk"}' \                          
  localhost:50052 auth.v1.AuthService.ValidateToken
{
  "userId": "98ac2ffe-5749-4ade-9053-6a7555d2bf04",
  "valid": true
}
```

## Gateway
### ping
```
curl -X GET "127.0.0.1:8080/ping" 
pong%
```

### Budget
```
curl -X POST "127.0.0.1:8080/api/budget/" -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiOThhYzJmZmUtNTc0OS00YWRlLTkwNTMtNmE3NTU1ZDJiZjA0IiwiZXhwIjoxNzY2MzQ1NTkwLCJpYXQiOjE3NjYyNTkxOTB9.HuEGJC4maaSVSNZwwFOksYHKo6B-THgAQoH-S8fItQk" -d '{"category": "test", "limit": 1000}'

curl -X GET "127.0.0.1:8080/api/budget/?category=test" -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiOThhYzJmZmUtNTc0OS00YWRlLTkwNTMtNmE3NTU1ZDJiZjA0IiwiZXhwIjoxNzY2MzQ1NTkwLCJpYXQiOjE3NjYyNTkxOTB9.HuEGJC4maaSVSNZwwFOksYHKo6B-THgAQoH-S8fItQk"
{"category":"test","limit":1000}% 

curl -X GET "127.0.0.1:8080/api/budget/list" -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiOThhYzJmZmUtNTc0OS00YWRlLTkwNTMtNmE3NTU1ZDJiZjA0IiwiZXhwIjoxNzY2MzQ1NTkwLCJpYXQiOjE3NjYyNTkxOTB9.HuEGJC4maaSVSNZwwFOksYHKo6B-THgAQoH-S8fItQk"
[{"category":"test","limit":1000}]% 
```

### Transactions
```
curl -X POST "127.0.0.1:8080/api/transactions/" -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiOThhYzJmZmUtNTc0OS00YWRlLTkwNTMtNmE3NTU1ZDJiZjA0IiwiZXhwIjoxNzY2MzQ1NTkwLCJpYXQiOjE3NjYyNTkxOTB9.HuEGJC4maaSVSNZwwFOksYHKo6B-THgAQoH-S8fItQk" -d '{"category": "test", "amount": 1, "date": "2025-12-10", "description": "test"}' 
{"id":2}% 

curl -X GET "127.0.0.1:8080/api/transactions/?id=2" -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiOThhYzJmZmUtNTc0OS00YWRlLTkwNTMtNmE3NTU1ZDJiZjA0IiwiZXhwIjoxNzY2MzQ1NTkwLCJpYXQiOjE3NjYyNTkxOTB9.HuEGJC4maaSVSNZwwFOksYHKo6B-THgAQoH-S8fItQk" 
{"id":2,"amount":1,"category":"test","description":"test","Date":"2025-12-10T00:00:00Z"}%  

curl -X GET "127.0.0.1:8080/api/transactions/list" -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiOThhYzJmZmUtNTc0OS00YWRlLTkwNTMtNmE3NTU1ZDJiZjA0IiwiZXhwIjoxNzY2MzQ1NTkwLCJpYXQiOjE3NjYyNTkxOTB9.HuEGJC4maaSVSNZwwFOksYHKo6B-THgAQoH-S8fItQk" 
[{"id":2,"amount":1,"category":"test","description":"test","Date":"2025-12-10T00:00:00Z"},{"id":1,"amount":1000,"category":"test","description":"test","Date":"2024-12-10T00:00:00Z"}]%  

curl -X GET "127.0.0.1:8080/api/transactions/export.csv" -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiOThhYzJmZmUtNTc0OS00YWRlLTkwNTMtNmE3NTU1ZDJiZjA0IiwiZXhwIjoxNzY2MzQ1NTkwLCJpYXQiOjE3NjYyNTkxOTB9.HuEGJC4maaSVSNZwwFOksYHKo6B-THgAQoH-S8fItQk" 
ID,Amount,Category,Description,Date
2,1.00,test,test,2025-12-10T00:00:00Z
1,1000.00,test,test,2024-12-10T00:00:00Z

curl -X POST "127.0.0.1:8080/api/transactions/bulk" -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiOThhYzJmZmUtNTc0OS00YWRlLTkwNTMtNmE3NTU1ZDJiZjA0IiwiZXhwIjoxNzY2MzQ1NTkwLCJpYXQiOjE3NjYyNTkxOTB9.HuEGJC4maaSVSNZwwFOksYHKo6B-THgAQoH-S8fItQk" -d '{"transactions": [{"category": "test2", "amount": 1, "date": "2025-12-10", "description": "test"}]}'
{"Accepted":1,"Rejected":0,"Errors":null}%
```

### Reports
```
curl -X GET "127.0.0.1:8080/api/reports/summary?from=2025-01-01&to=2025-12-31" -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiOThhYzJmZmUtNTc0OS00YWRlLTkwNTMtNmE3NTU1ZDJiZjA0IiwiZXhwIjoxNzY2MzQ1NTkwLCJpYXQiOjE3NjYyNTkxOTB9.HuEGJC4maaSVSNZwwFOksYHKo6B-THgAQoH-S8fItQk" 
{"report":{"test":1,"test2":1},"cache_result":false}%  
```

### Тесты
```
cd ledger && go test ./... -cover -coverprofile=cover.out && go tool cover -html=cover.out -o cover.html
cd gateway && go test ./... -cover -coverprofile=cover.out && go tool cover -html=cover.out -o cover.html
```