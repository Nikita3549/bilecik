FROM golang:1.26.4 as build
WORKDIR /opt/api

COPY go.mod go.sum ./

RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

COPY ./cmd ./cmd
COPY ./internal ./internal
COPY ./pkg ./pkg

RUN go build ./cmd/main.go

FROM alpine:latest
WORKDIR /opt/api

COPY ./migrations ./migrations/
COPY --from=build /go/bin/migrate /usr/local/bin/migrate
COPY --from=build /opt/api/main /opt/api/main

CMD . ./.env \
  && migrate -path migrations -database "postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable" up \
  && exec ./main
