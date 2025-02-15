# wasm
FROM golang:alpine AS wasm

WORKDIR /src

COPY wasm/go.* ./
RUN go mod download -x

COPY wasm/ ./
RUN GOOS=js GOARCH=wasm go build -o web/app.wasm
RUN cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" web/

# svelte
FROM oven/bun:alpine AS svelte

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
COPY --from=wasm /src/web web/
COPY --from=svelte /app/build ui/

EXPOSE 1106

ENTRYPOINT ["./shorty"]
