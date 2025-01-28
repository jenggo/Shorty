FROM oven/bun:alpine

ARG VITE_API_BASE_URL=u.nusatek.dev

WORKDIR /app

COPY ui/package.json ui/bun.lockb ./
RUN bun install

COPY ui/ ./
RUN bun run lint
RUN bun run build

# do not use scratch, uploading will failed if file >= 20MB
FROM chainguard/static

COPY shorty config.yaml ./
COPY --from=0 /app/build ui/

EXPOSE 1106

ENTRYPOINT ["./shorty"]
