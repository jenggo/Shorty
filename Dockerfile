FROM oven/bun:alpine

ARG API_URL=u.nusatek.dev

WORKDIR /app

COPY ui/package.json ui/bun.lockb ./
RUN bun install

COPY ui/ ./
RUN VITE_API_BASE_URL=$API_URL bun run build

RUN apk add --update ca-certificates && addgroup -S groot && adduser -S groot -G groot

FROM scratch
ARG APP_NAME=shorty

WORKDIR /app

COPY ${APP_NAME} config.yaml ./
COPY --from=0 /etc/passwd /etc/passwd
COPY --from=0 /app/build ui/

EXPOSE 1106

USER groot

ENTRYPOINT ["./shorty"]
