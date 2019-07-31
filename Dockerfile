# build stage
FROM golang:1.12.7-alpine3.10 AS build-env
RUN apk update && apk add gcc=8.3.0-r0 g++=8.3.0-r0 git=2.22.0-r0
WORKDIR /src
COPY . . 
RUN go mod download
RUN go mod verify
RUN export CGO_ENABLED=1 && go build -o /src/bin/sf

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
