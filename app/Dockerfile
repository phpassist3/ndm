# Multi-stage сборка: статичный бинарь в первом слое, минимальный рантайм во втором.
# Итоговый образ — несколько мегабайт, поднимается мгновенно, поверхность атаки минимальна.

FROM golang:1.23-alpine AS build
WORKDIR /src

# Сначала только go.mod — слой кэшируется и не пересобирается при изменениях кода.
COPY go.mod ./
RUN go mod download

# Теперь сам код.
COPY app.go ./

# CGO выключаем, чтобы получить полностью статичный бинарь, который запустится
# в scratch/distroless/alpine без зависимостей от libc.
ENV CGO_ENABLED=0 GOOS=linux
RUN go build -trimpath -ldflags="-s -w" -o /out/app .


FROM alpine:3.20

# Запускаем не от root — стандартная защитная практика.
RUN addgroup -S app && adduser -S -G app app
USER app

COPY --from=build /out/app /usr/local/bin/app

EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/app"]
