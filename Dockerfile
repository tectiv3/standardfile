# build stage
FROM golang:alpine AS build-env
RUN apk update && apk --no-cache add gcc g++ git
WORKDIR /src
RUN git clone --branch master --depth 1 https://github.com/tectiv3/standardfile.git .

ENV GOPROXY=https://proxy.golang.org
RUN go mod download
RUN go mod verify

RUN SF_VERSION=$(git describe --tags) \
    && BUILD_TIME=`date +%FT%T%z` \
    && go build -ldflags="-w -X main.BuildTime=${BUILD_TIME} -X main.Version=${SF_VERSION}" -o /src/bin/sf
# RUN /src/bin/sf -v

# final stage
FROM alpine

RUN mkdir -p /data

WORKDIR /app
COPY --from=build-env /src/bin/sf /app/sf

VOLUME /data
EXPOSE 8888

ENTRYPOINT [ "/app/sf" ]
CMD [ "-c", "/data", "--foreground", "-db", "/data/sf.db"]
