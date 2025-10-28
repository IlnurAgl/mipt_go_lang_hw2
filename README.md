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