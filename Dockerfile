ARG USER_CREATE_IMAGE=alpine
ARG DEPLOY_IMAGE=scratch

FROM ${USER_CREATE_IMAGE}

RUN addgroup -S groot && adduser -S groot -G groot


FROM ${DEPLOY_IMAGE}
ARG APP_NAME=shorty

WORKDIR /app

COPY ${APP_NAME} config.yaml ./
COPY ui/ ui/
COPY --from=0 /etc/passwd /etc/passwd

EXPOSE 1106

USER groot

ENTRYPOINT ["./shorty"]
