FROM alpine:latest
RUN apk add --no-cache file
WORKDIR /app
COPY ./main ./config.yaml .
EXPOSE 6543
CMD ["./main"]