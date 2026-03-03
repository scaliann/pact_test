FROM golang:1.26 AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/app ./cmd/app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/ui ./cmd/ui

FROM gcr.io/distroless/static-debian12 AS app-runtime
WORKDIR /app
COPY --from=build /out/app /app/app
EXPOSE 50051
ENTRYPOINT ["/app/app"]

FROM gcr.io/distroless/static-debian12 AS ui-runtime
WORKDIR /app
COPY --from=build /out/ui /app/ui
EXPOSE 8085
ENTRYPOINT ["/app/ui"]
