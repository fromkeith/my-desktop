(WIP) A Gmail frontend that puts productivity first.

- Have multiple emails open at the same time
- Read what you are replying to, next to your reply
- Stop wasting all the UI real estate on an a list of emails
- Plug in AI agents to organize, automate and prioritize your emails

This is an active work in progress.


# Dev Setup


## Server

(Go, Postgres, MongoDB, Kafka)

Uses [Task](https://taskfile.dev/) to build and run.

1. Prepare docker to allow Postgres, MongoDB, and Kafka to run

```
to your hosts file add:
> 127.0.0.1 devbox.local
```

This lets MongoDB (with vector searching) to talk to each other.

2. Starup docker

```bash
cd server
docker compose up -d

# (or use task)
task up
```

3. Setup the Postgres. You will need [dbmate](https://github.com/amacneil/dbmate) installed already.

```bash
task migrate-postgres
```

4. Setup MongoDB. You will need [MigrateMongo](https://www.npmjs.com/package/migrate-mongo) installed.

```bash
task migrate-mongo
```

5. Get dependancies for go. Make sure you have gcc on your path (eg. C:\msys64\mingw64\bin)

```bash
go mod download
```

6. Setup env files (see below)

7. Build and run

```bash
task run-all
```

The main server will listen on 8080. However, the front-end (vite) will automatically proxy any `localhost:5173/api` calls to the server.

## Front-end

(Vite, Svelte, Tailwind)

1. Build the server Swagger definition.

```bash
cd server
go install github.com/swaggo/swag/cmd/swag@latest
task swag
```

2. Build the client side side Typescript models from the Swagger def.

```bash
npm run models
```

3. Build and run the client:

```bash
npm install
npm run dev
```

It will start vite listening on: `http://localhost:5173`. You can now open your browser to that.

## Env files

`server/.env`

```
# access to gmail + gemini
GOOGLE_API_KEY=
GOOGLE_CLIENT_ID=
GOOGLE_SECRET_ID=
REDIRECT_URI=
GOOGLE_CREDENTIALS=
# random secret key for signing JWT tokens
JWT_KEY=
# these are just local values... they should change for prod
DOMAIN=localhost:5173
DOMAIN_URL=http://localhost:5173/
MONGODB_URI=mongodb://devbox.local:27017/?replicaSet=rs0
MONGODB_DB=MyDesktop
KAFKA_URI=localhost:9092
POSTGRES_URI=postgres://postgres:postgres@localhost:5432/desktop?sslmode=disable
```


# Troubleshooting 

### Building

```
Some people have good experiences with MSYS2: https://www.msys2.org/. After installing MSYS2, run pacman -S mingw-w64-x86_64-toolchain mingw-w64-x86_64-pkg-config to install MinGW and pkg-config. (This is the most recommended way by now.)
```


# Architecture

## Code Layout

- `src/` - front-end code
    - `lib/db` - [RXDB](https://rxdb.info/) data sync base. We use [Dexie](https://dexie.org/) for storage.
    - `lib/models` - any Typescript models
    - `lib/components/ui` - [ShadCn](https://shadcn-svelte.com/docs) components
    - `lib/my-components` - our custom components
    - `lib/pods` - [SvelteProvider](https://github.com/fromkeith/SvelteProvider) pods for state/data pulling
    
- `server/` - server code
    - `main.go` - The main :)
    - `services/` - A series of background services, heavily relies on Kafka.
        - `email-injestor` - Pulls new email content from GMail and saves it to MongoDB.
        - `gemini` - Triggered off a new email. Runs the email through Gemini, and saves the vectors, categories, and tags to MongoDB.
        - `tagsAndCats` - Listens to MongoDB "Messages". Makes categories and tags searchable + keeps a counter for each account.
        - `messageToThread` - Listens to MongoDB "Messages". Puts messages into "MessageThreads" collection.

# TODO

[] Lots!
[] Make JWT tokens more secure.. currently its just a "get running" mode...
