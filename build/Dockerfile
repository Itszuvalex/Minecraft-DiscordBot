ARG ExePath
ARG ExeNameBuilt
ARG ExeName

FROM alpine:latest

# Necessary to connect to things on the web
RUN apk update && apk add ca-certificates

COPY $ExePath/$ExeNameBuilt /srv/mcdiscord/$ExeName

ENTRYPOINT ["/srv/mcdiscord/$ExeName"]
CMD [""]
