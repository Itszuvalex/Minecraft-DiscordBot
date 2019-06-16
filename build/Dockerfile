FROM alpine:latest

# Necessary to connect to things on the web
RUN apk update && apk add ca-certificates

COPY mcdiscord_unix /srv/mcdiscord/mcdiscord

ENTRYPOINT ["/srv/mcdiscord/mcdiscord"]
CMD [""]
