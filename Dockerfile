FROM alpine:latest

COPY mcdiscord.exe /srv/mcdiscord/

# Necessary to connect to things on the web
RUN apk update && apk add ca-certificates

ENTRYPOINT ["/srv/mcdiscord/mcdiscord.exe"]
CMD [""]
