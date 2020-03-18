FROM golang:1.14.0-alpine as build-env
WORKDIR /customer
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main main.go

FROM alpine:latest
WORKDIR /app
ARG POSTGRESQL_URL
ARG POSTGRESQL_USERNAME
ARG POSTGRESQL_PASSWORD
ARG REDIS_URL
ARG ZIPKIN_URL
ARG SERVICE
ARG NAMESPACE
ARG GRPCADDR
ENV POSTGRESQL_URL=${POSTGRESQL_URL}
ENV POSTGRESQL_USERNAME=${POSTGRESQL_USERNAME}
ENV POSTGRESQL_PASSWORD=${POSTGRESQL_PASSWORD}
ENV REDIS_URL=${REDIS_URL}
ENV ZIPKIN_URL=${ZIPKIN_URL}
ENV SERVICE=${SERVICE}
ENV NAMESPACE=${NAMESPACE}
ENV GRPCADDR=${GRPCADDR}
COPY --from=build-env /customer/internal/config/dev.yml ./internal/config/dev.yml
COPY --from=build-env /customer/main .
# COPY ./internal/config/dev.yml ./internal/config/dev.yml
# COPY ./build/main .
EXPOSE 9999
# RUN ./main -zipkinAddr ${ZIPKIN_URL} -dbHost ${POSTGRESQL_URL} -dbUserName ${POSTGRESQL_USERNAME} -dbPassword ${POSTGRESQL_PASSWORD} -redisAddr ${REDIS_URL}
# ENTRYPOINT [ "./main", "-zipkinAddr", "${ZIPKIN_URL}", "-dbHost", "${POSTGRESQL_URL}", "-dbUserName", "${POSTGRESQL_USERNAME}", "-dbPassword", "${POSTGRESQL_PASSWORD}", "-redisAddr", "${REDIS_URL}"]
ENTRYPOINT [ "/bin/sh","-c", "./main -zipkinAddr=${ZIPKIN_URL} -dbHost=${POSTGRESQL_URL} -dbUserName=${POSTGRESQL_USERNAME} -dbPassword=${POSTGRESQL_PASSWORD} -redisAddr=${REDIS_URL} -service=${SERVICE} -namespace=${NAMESPACE} -grpcAddr=${GRPCADDR}"]
