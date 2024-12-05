FROM docker.m.daocloud.io/alpine:3.15 AS artifacts

ARG APP
ARG TARGETOS
ARG TARGETARCH

COPY out/$TARGETOS/$TARGETARCH/$APP /files/app/${APP}

FROM docker.m.daocloud.io/alpine:3.15

WORKDIR /app

ARG APP
ARG TARGETOS
ARG TARGETARCH

ENV APP ${APP}

ARG VERSION
ENV VERSION ${VERSION}

COPY --from=artifacts /files /

CMD /app/${APP}
