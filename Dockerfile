FROM golang:1.26.1

ARG GOBUSTER_VERSION=v3.8.2

ENV DEBIAN_FRONTEND=noninteractive
ENV PATH="/go/bin:/usr/local/go/bin:${PATH}"

WORKDIR /app

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go install github.com/OJ/gobuster/v3@${GOBUSTER_VERSION} \
    && go build -o /usr/local/bin/burstui .

CMD ["burstui"]
