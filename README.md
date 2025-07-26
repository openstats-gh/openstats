# openstats

Stats & achievement tracking website for games. Follow other players & compare your stats. Showcase your achievement 
progress, games you've 100%'d, your rarest achievements, and more!

## Who is this for?

### For players

Player profiles, achievement showcases, and more.

### For game developers

openstats has a simple webapi for developers. Developers can track & update achievement progress, and log statistics 
such as playtime.

### For me

I'm tired of being locked into a proprietary game platform. It feels like Steam is the only platform that does 
achievements & stats somewhat right. It was able to achieve that through its monopolistic saturation as a gaming social 
network. I don't want Steam to be the only choice players have for simple things like achievement tracking and profile 
showcases.

## Hosting

WIP! Come back some time later™ and I'll have hopefully updated this to include more concrete self-hosting 
instructions.

## In the wild

Soon...

## Development

### Setup

1. Install [docker](https://docs.docker.com/engine/install/) & https://docs.docker.com/compose/install/
2. Install go 1.24
3. Install node.js 24 & npm 11
4. Install `migrate`
    ```shell
    go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.3
    ```
5. Install web dependencies
    ```shell
    cd web
    npm i
    ```
6. Create `api/env.local` and `web/env.local`
   - See `api/env.example` and `web/env.local` for instructions

> [!NOTE]
> If you see something like `command not found` when trying to use `migrate`, chances are the gopath `go/bin` directory isn't on your `PATH`! This is usually located in your home directory e.g. `C:/Users/YourUserName/go/bin` or `/home/username/go/bin`. See `go help install` for more information.

Further reading:

- `go help install`, `go help build`, `go help run`
- [A TUI for docker](https://github.com/jesseduffield/lazydocker)
- [Docker CLI cheatsheet](https://docs.docker.com/get-started/docker_cheatsheet.pdf)
- [Docker Compose manual](https://docs.docker.com/compose/)
- [Fiber backend web framework](https://gofiber.io)
- [Svelte & SvelteKit frontend framework](https://svelte.dev/)

### Start/stop local postgres db & pgadmin

In `api` as current working directory.

Starting:

```shell
docker compose up -d
```

Stopping:

```shell
docker compose down
```

The local db is accessible at `postgres://openstats:openstats@localhost:15432/openstats?sslmode=disable`

The local pgadmin webserver is accessible at http://localhost:15433

### Start API server

Expects the postgres database to be alive. See above.

In `api` as current working directory.

```shell
go run
```

I recommend using an IDE with Go debugging integration such as VS Code or Jetbrains Goland, and setting up
a run & debug configuration.

### Start frontend server

Expects the API to be alive. See above.

In `web` as current working directory.

```shell
npm run dev
```

### Create a migration

Its fine to test changes to the database schema ad-hoc without creating a migration. However, if you intend to commit
your changes, you must create a migration:

```shell
migrate create -ext sql -dir migrations a-summary-of-your-changes
```

### Run migrations 

```shell
migrate -source file://migrations -database postgres://openstats:openstats@localhost:15432/openstats?sslmode=disable u
```

### pgadmin

`docker-compose.yml` is configured to run a pgadmin web server next to the postgres server. It can be accessed at 
http://localhost:15433.
