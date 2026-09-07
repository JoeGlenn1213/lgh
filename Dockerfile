# LGH MCP server container — used by MCP registries (e.g. Glama) to start
# the server and run protocol introspection checks. `lgh mcp` exposes the
# 10 repo/server-control tools over stdio.
FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/lgh ./cmd/lgh

FROM alpine:3.20
RUN apk add --no-cache git
COPY --from=build /out/lgh /usr/local/bin/lgh
ENTRYPOINT ["lgh", "mcp"]
