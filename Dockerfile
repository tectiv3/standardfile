# build stage
FROM golang:alpine AS build-env
RUN apk update \
    && apk --no-cache add gcc g++ git
WORKDIR /src
COPY . .

# ENV GO111MODULE=on
# RUN go mod download
# RUN go mod verify

RUN SF_VERSION=$(git describe --tags) \
    && BUILD_TIME=`date +%FT%T%z` \
    && CGO_ENABLED=1 \
    && go build -mod=vendor -ldflags="-w -X main.BuildTime=${BUILD_TIME} -X main.Version=${SF_VERSION}" -o /src/bin/sf
# RUN /src/bin/sf -v

# final stage
FROM alpine

RUN mkdir -p /data

WORKDIR /app
COPY --from=build-env /src/bin/sf /app/sf

VOLUME /data
EXPOSE 8888
ENTRYPOINT ["/app/sf", "-c", "/data", "--foreground", "-db", "/data/sf.db"]