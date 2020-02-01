FROM alpine:3.11

RUN mkdir -p /app/bin
WORKDIR /app

COPY ./bin /app/bin
COPY config.yml /app
COPY ./html /app/html

RUN addgroup -g 1001 worker && \
    adduser --system --uid 1001 worker worker && \
    chown -R worker:worker /app

EXPOSE 8080
USER worker

CMD ./bin/app
