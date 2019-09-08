FROM golang:alpine AS build-env
ADD . /src
RUN cd /src && go build -o app

# final stage
FROM alpine
WORKDIR /tg-retard-jokes-bot
COPY --from=build-env /src/app /tg-retard-jokes-bot/
ENTRYPOINT ./app