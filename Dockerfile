# --- build ---
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN go build -o /out/frostload ./cmd/frostload
RUN go build -o /out/backend ./cmd/backend

#--- run stage
FROM alpine:3.20
COPY --from=build /out/frostload /usr/local/bin/frostload
COPY --from=build /out/backend /usr/local/bin/backend
