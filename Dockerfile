ARG BASE_IMAGE=alpine
# ARG BASE_IMAGE=golang:alpine
ARG DEPLOY_IMAGE=scratch

FROM ${BASE_IMAGE}
ARG APP_NAME=shorty

RUN apk add --update ca-certificates tzdata
# RUN apk add --update upx dumb-init tzdata && wget -O- -nv https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s

# WORKDIR /src
# COPY go.* ./
# RUN go mod download -x

# COPY . .
# RUN golangci-lint run
# RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o ${APP_NAME} -trimpath && upx -q --best --lzma ${APP_NAME}

FROM ${DEPLOY_IMAGE}
ARG APP_NAME=shorty

WORKDIR /app

COPY ${APP_NAME} .
COPY --from=0 /etc/ssl/certs /etc/ssl/certs
COPY --from=0 /usr/share/zoneinfo /usr/share/zoneinfo
COPY config.yaml .

EXPOSE 1106

ENTRYPOINT ["./shorty"]
