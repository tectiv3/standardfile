# build stage
FROM alpine:3.10 AS build-env
RUN apk update
RUN apk upgrade
RUN apk add --update go=1.12.6-r0 gcc=8.3.0-r0 g++=8.3.0-r0 git=2.22.0-r0
WORKDIR /src
#ENV GOPATH /src
COPY . . 
RUN go mod download
RUN go mod verify
#RUN go get -d -v
RUN export CGO_ENABLED=1 && go build -o /src/bin/sf

#RUN go get server # server is name of our application
#RUN CGO_ENABLED=1 GOOS=linux go install -a server

# final stage
FROM alpine
RUN apk update && apk add --no-cache sqlite
RUN mkdir -p /stdfile/db
WORKDIR /app
COPY --from=build-env /src/bin/sf /app/sf
COPY --from=build-env /src/standardfile.json /app/standardfile.json
VOLUME /stdfile/db
EXPOSE 8888
ENTRYPOINT ["/app/sf"]
