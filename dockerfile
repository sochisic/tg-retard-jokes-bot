FROM golang:1.13 AS build

WORKDIR /go/src/bot
ADD . .
RUN go get -d ./... 
RUN go install -v ./...
COPY . /go/src/bot
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/bot

# final stage
FROM golang:1.13-alpine
WORKDIR /bot
COPY --from=build /bin/bot /bin/bot
EXPOSE 80
EXPOSE 443
ENTRYPOINT ["/bin/bot"]