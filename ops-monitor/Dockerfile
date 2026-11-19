FROM alpine:3.21
RUN apk add --no-cache curl
WORKDIR /monitor
COPY health-check.sh ./health-check.sh
CMD ["sh", "./health-check.sh"]
