FROM alpine:latest

COPY mcdiscord /srv/mcdiscord/

# Necessary to connect to things on the web
RUN apk update && apk add ca-certificates

ENTRYPOINT ["/srv/mcdiscord/mcdiscord"]
CMD [""]
