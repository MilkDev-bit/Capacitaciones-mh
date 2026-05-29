module Prueba-Go/services/auth

go 1.26.2

require (
	Prueba-Go/gen v0.0.0
	github.com/golang-jwt/jwt/v5 v5.3.1
	github.com/google/uuid v1.6.0
	github.com/jmoiron/sqlx v1.4.0
	github.com/lib/pq v1.12.3
	golang.org/x/crypto v0.50.0
	google.golang.org/grpc v1.70.0
)

require (
	golang.org/x/net v0.52.0 // indirect
	golang.org/x/sys v0.43.0 // indirect
	golang.org/x/text v0.36.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241202173237-19429a94021a // indirect
	google.golang.org/protobuf v1.36.10 // indirect
)

replace Prueba-Go/gen => ../../gen
