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

In `api` as current working directory.

Its fine to test changes to the database schema ad-hoc without creating a migration. However, if you intend to commit
your changes, you must create a migration:

```shell
migrate create -ext sql -dir db/migrations a-summary-of-your-changes
```

After writing your DDL for the new migration, regenerate Go models based on your changes:

```shell
go generate
```

See [Add a SQL Query](#add-a-sql-query) for more information on how we generate models & queries.

### Run migrations 

In `api` as current working directory.

```shell
migrate -source file://db/migrations -database postgres://openstats:openstats@localhost:15432/openstats?sslmode=disable u
```

### Add a SQL Query

We use [sqlc](https://sqlc.dev) to generate structs and functions for each table and query. sqlc is configured in [api/sqlc.yaml](./api/sqlc.yaml) to look for migrations at [api/db/migrations](./api/db/migrations/), and for queries at [api/db/sql](./api/db/sql).

Each sql file may hold many queries, and each query is separated by a comment like this:

```sql
-- name: FindUser :one
select * from users where id = $1 limit 1;

-- name: FindUserBySlug :one
select * from users where id = $1 limit 1;
```

To generate the Go code after making changes or after adding a new query, run this with `api` as your working directory:

```shell
go generate
```

See the [sqlc docs](https://docs.sqlc.dev/en/v1.29.0/) for more information on query annotations, parameterization, etc.

### Using `sqlc`-generated code

Each table gets its own struct, and each query gets its own function. If the function is parameterized with multiple parameters, then it'll get a params struct.

Given this query:

```sql
-- name: AddOrUpdateAchievement :one
insert into achievement (game_id, slug, name, description, progress_requirement)
values ($1, $2, $3, $4, $5)
on conflict(game_id, slug)
    do update set name=excluded.name,
                  description=excluded.description,
                  progress_requirement=excluded.progress_requirement
returning case when achievement.created_at == achievement.updated_at then true else false end as is_new;
```

Usage might look like this:

```go
isNew, createErr := Queries.AddOrUpdateAchievement(ctx.Context(), query.AddOrUpdateAchievementParams{
    GameID:              game.ID,
    Slug:                achievementSlug,
    Name:                request.Name,
    Description:         request.Description,
    ProgressRequirement: request.ProgressRequirement,
})

if createErr != nil {
    log.Error(createErr)
    return ctx.SendStatus(fiber.StatusInternalServerError)
}

if isNew {
    newLocation, routeErr := ctx.GetRouteURL("readAchievement", fiber.Map{"devSlug": devSlug, "gameSlug": gameSlug, "achievementSlug": achievementSlug})
    if routeErr == nil {
        ctx.Location(newLocation)
    }
    return ctx.SendStatus(fiber.StatusCreated)
}
```

If a query returns an entire table, e.g.:

```sql
-- name: FindUser :one
select * from users where id = $1 limit 1;
```

Then a fully type-qualified usage would look like this:

```go
import (
	"github.com/dresswithpockets/openstats/app/db/query"
)

// ...
    var user query.User, err error = Queries.FindUser(c.Context(), userId)
```
