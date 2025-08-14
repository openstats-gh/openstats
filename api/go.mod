module github.com/dresswithpockets/openstats/app

go 1.24.2

toolchain go1.24.5

require (
	github.com/Masterminds/squirrel v1.5.4
	github.com/danielgtaylor/huma/v2 v2.34.1
	github.com/eknkc/basex v1.0.1
	github.com/go-chi/chi/v5 v5.2.2
	github.com/go-chi/httplog/v3 v3.2.2
	github.com/go-playground/validator/v10 v10.27.0
	github.com/gofiber/fiber/v2 v2.52.8
	github.com/gofiber/storage/postgres/v3 v3.2.0
	github.com/golang-jwt/jwt/v5 v5.2.3
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.7.5
	github.com/rotisserie/eris v0.5.4
	github.com/rs/cors v1.11.1
	github.com/vgarvardt/pgx-google-uuid/v5 v5.6.0
	golang.org/x/crypto v0.40.0
)

require (
	github.com/andybalholm/brotli v1.1.1 // indirect
	github.com/gabriel-vasile/mimetype v1.4.9 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/lann/builder v0.0.0-20180802200727-47ae307949d0 // indirect
	github.com/lann/ps v0.0.0-20150810152359-62de8c46ede0 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.16 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.62.0 // indirect
	golang.org/x/net v0.42.0 // indirect
	golang.org/x/sync v0.16.0 // indirect
	golang.org/x/sys v0.34.0 // indirect
	golang.org/x/text v0.27.0 // indirect
)

replace github.com/mattn/go-sqlite3 => github.com/dresswithpockets/go-sqlite3 v1.14.28-2
