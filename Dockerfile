# Create Builder Image
FROM --platform=linux/amd64 golang:1.23.8 as builder
LABEL maintainer="DPS <adityakurnia.p@gmail.com>"

ENV GIT_TERMINAL_PROMPT=1 GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=amd64

# Set Working Directory
RUN mkdir -p /app
ADD . /app
WORKDIR /app
COPY . .

# Do Your Magic Here
#
#

# Build Go Binary File
RUN GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/main

# Create Second Image
FROM --platform=linux/amd64 alpine:3.13.1

RUN touch .env
ENV TZ=Asia/Jakarta

RUN apk add --no-cache tzdata
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

RUN mkdir -p /tempfile
RUN mkdir -p /assets

# Copy Binary File from Builder Image
COPY --from=builder /app/main /main
COPY --from=builder /app/.env /.ENV
COPY --from=builder /app/static /static

# Run Binary File
ENTRYPOINT ["/main"]
