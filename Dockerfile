FROM oven/bun:alpine

ARG VITE_API_BASE_URL=u.nusatek.dev

RUN addgroup -S groot && adduser -S groot -G groot
RUN apk add --update ca-certificates

WORKDIR /app

COPY ui/package.json ui/bun.lockb ./
RUN bun install

COPY ui/ ./
RUN bun run lint
RUN bun run build

# do not use scratch, uploading will failed if file >= 20MB
FROM busybox:musl

WORKDIR /app

COPY shorty config.yaml ./
COPY --from=0 /etc/passwd /etc/passwd
COPY --from=0 /app/build ui/

EXPOSE 1106

USER groot

ENTRYPOINT ["./shorty"]
