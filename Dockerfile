FROM --platform=$BUILDPLATFORM node:22-alpine AS builder-web
WORKDIR /app
COPY ./gui/package.json ./gui/yarn.lock ./
RUN yarn install
COPY ./gui .
RUN echo "network-timeout 600000" >> .yarnrc
RUN yarn config set registry https://registry.npm.taobao.org
RUN yarn config set sass_binary_site https://cdn.npm.taobao.org/dist/node-sass -g
RUN yarn cache clean && yarn && yarn build

FROM --platform=$BUILDPLATFORM golang:trixie AS builder
WORKDIR /app
# Define the build arguments passed from GitHub Actions
ARG APP_VERSION=v0.0.0
ARG APP_COMMIT=unknown
RUN set -eux;     \
    apt update -y; \
    apt install -y --no-install-recommends       \
        ca-certificates       \
        build-essential;     \
    apt-mark showmanual > /savedAptMark.txt
RUN set -eux;   \
    apt-mark auto '.*' > /dev/null ;	\
    apt-mark manual $(cat /savedAptMark.txt) > /dev/null; 	\
    apt purge -y --auto-remove -o APT::AutoRemove::RecommendsImportant=false;     \
    apt clean;     \
    apt autoclean;     \
    rm -rf /var/lib/apt/lists/*
COPY ./service/go.mod ./service/go.sum ./
RUN set -eux;   \
    go mod download
COPY ./service .
COPY --from=builder-web /app/web ./server/router/web
RUN set -eux;   \
    export BUILD_DATE="$(date +%Y-%m-%d)";   \
    CGO_ENABLED=1 \
        GOOS=linux \
        go build \
            -buildvcs=false \
            -ldflags="-s -w -X 'main.Version=${APP_VERSION}' -X 'main.Commit=${APP_COMMIT}' -X 'main.BuildDate=${BUILD_DATE}'" \
            -o ./v2raya . ;  \
    chmod +x ./v2raya

FROM ghcr.io/v2fly/v2ray:v5.53-extra
COPY --from=builder /build/service/v2raya /usr/bin/
ENV TZ="Asia/Jakarta"
RUN [ ! -f /etc/localtime ] && ln -s /usr/share/zoneinfo/\$TZ /etc/localtime; \
    echo \$TZ > /etc/timezone
RUN set -eux;     \
    apt update -y; \
    apt install -y --no-install-recommends       \
        ca-certificates       \
        build-essential       \
        wget;     \
    apt-mark showmanual > /savedAptMark.txt
RUN set -eux;     \
    wget -O /usr/local/share/v2ray/LoyalsoldierSite.dat https://raw.githubusercontent.com/mzz2017/dist-v2ray-rules-dat/master/geosite.dat
RUN set -eux;   \
    apt-mark auto '.*' > /dev/null ;	\
    apt-mark manual $(cat /savedAptMark.txt) > /dev/null; 	\
    apt purge -y --auto-remove -o APT::AutoRemove::RecommendsImportant=false;     \
    apt clean;     \
    apt autoclean;     \
    rm -rf /var/lib/apt/lists/*
EXPOSE 2017
VOLUME /etc/v2raya
ENTRYPOINT ["v2raya"]
