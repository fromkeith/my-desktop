

# Server

Listens on port 8080
Vite (npm run dev) will automatically proxy any /api call to the server

## Database

Migrations

Use: https://github.com/amacneil/dbmate

```bash
dbmate -d .\migrations -u sqlite:desktop.sqlite3 up
```

## Webhooks

```
npm install hookdeck-cli -g
hookdeck login
hookdeck listen 5173 goauth
```


## Building

Use build.bat... it sets the needed cgo flags. Make sure you have gcc:

```
Some people have good experiences with MSYS2: https://www.msys2.org/. After installing MSYS2, run pacman -S mingw-w64-x86_64-toolchain mingw-w64-x86_64-pkg-config to install MinGW and pkg-config. (This is the most recommended way by now.)
```
