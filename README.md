

# Server

Listens on port 8080
Vite (npm run dev) will automatically proxy any /api call to the server

## Database (Postgres)

Migrations

Use: https://github.com/amacneil/dbmate

```bash
dbmate -d .\migrations -u postgres://postgres:postgres@localhost:5432/desktop?sslmode=disable up
```

## Database (MongoDB)

```
docker compose up -d
```

Migrations & db creation use: https://www.npmjs.com/package/migrate-mongo

```
cd server/migrations/mongo
migrate-mongo up
```



## Webhooks

```
npm install hookdeck-cli -g
hookdeck login
hookdeck listen 5173 goauth
```


## Building

Make sure you have gcc on your path (eg. C:\msys64\mingw64\bin)

```
Some people have good experiences with MSYS2: https://www.msys2.org/. After installing MSYS2, run pacman -S mingw-w64-x86_64-toolchain mingw-w64-x86_64-pkg-config to install MinGW and pkg-config. (This is the most recommended way by now.)
```

Uses https://taskfile.dev/docs/guide

```
task run-all
```


alternatively:




## Generating Swagger

```
go install github.com/swaggo/swag/cmd/swag@latest
swag init --parseVendor --parseDependency
```

client side
```
npm run models
```
