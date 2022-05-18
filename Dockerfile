FROM golang:1.17-alpine AS build_base

RUN apk add --no-cache git

WORKDIR /calc

COPY go.mod .

RUN go mod download

COPY . .

WORKDIR /calc/cmd/

RUN CGO_ENABLED=0 go test -v

RUN go build -o ./out/app .

FROM alpine:latest

RUN apk add ca-certificates

COPY --from=build_base calc/cmd/out/app /service/app
COPY --from=build_base calc/common/config.yml /service/config.yml
COPY --from=build_base calc/public.pem /service/public.pem
COPY --from=build_base calc/private.pem /service/private.pem
COPY docker-entrypoint.sh /service/docker-entrypoint.sh

EXPOSE 8080
EXPOSE 8081

RUN chmod 755 /service/docker-entrypoint.sh
ENTRYPOINT ["/service/docker-entrypoint.sh", "-c", "service/config.yml"]