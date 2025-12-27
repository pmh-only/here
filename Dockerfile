FROM scratch

ARG TARGETARCH
ARG user=1000
ARG group=1000

USER $user:$group
WORKDIR /app

COPY main-${TARGETARCH} ./main

ENTRYPOINT ["/app/main"]
