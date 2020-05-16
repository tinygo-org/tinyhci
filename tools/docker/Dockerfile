FROM debian:stable-slim AS tinygohci-base

RUN apt-get clean && apt-get update && \
    apt-get install -y wget gcc gcc-avr avr-libc git

ENV GO_RELEASE=1.14.2
RUN wget https://dl.google.com/go/go${GO_RELEASE}.linux-amd64.tar.gz && \
    tar xfv go${GO_RELEASE}.linux-amd64.tar.gz -C /usr/local && \
    rm go${GO_RELEASE}.linux-amd64.tar.gz
ENV PATH=${PATH}:/usr/local/go/bin

FROM tinygohci-base AS tinygohci-build
ARG TINYGO_DOWNLOAD_URL=https://github.com/tinygo-org/tinygo/releases/download/v0.13.1/tinygo0.13.1.linux-amd64.tar.gz
RUN wget ${TINYGO_DOWNLOAD_URL} -O tinygo.tar.gz && \
    tar -xzf tinygo.tar.gz -C /usr/local && \
    rm tinygo.tar.gz
ENV PATH=${PATH}:/usr/local/tinygo/bin
RUN go get -d tinygo.org/x/drivers

RUN apt-get remove -y wget && \
    apt-get autoremove -y && \
    apt-get clean

CMD ["tinygo"]
