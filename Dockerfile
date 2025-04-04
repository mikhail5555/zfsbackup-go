FROM alpine:3.21

RUN apk --no-cache add zfs

WORKDIR /app
COPY zfsbackup-go .

ENTRYPOINT [ "/app/zfsbackup-go" ]