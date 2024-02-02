ARG DEPLOY_IMAGE=scratch
FROM ${DEPLOY_IMAGE}
ARG APP_NAME=shorty

WORKDIR /app

COPY ${APP_NAME} config.yaml ./
COPY ui/ ui/

EXPOSE 1106

ENTRYPOINT ["./shorty"]
