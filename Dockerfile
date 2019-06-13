FROM alpine:latest

COPY mcdiscord.exe /srv/mcdiscord/

ENTRYPOINT ["/srv/mcdiscord/mcdiscord.exe"]
CMD [""]
