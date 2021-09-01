FROM debian:stable-slim AS tinygohci-base

RUN apt-get clean && apt-get update && \
    apt-get install -y wget gcc gcc-avr avr-libc avrdude git make build-essential libreadline-dev libwxgtk3.0-* python3 python3-pip openocd

RUN git clone https://github.com/shumatech/BOSSA.git && \
    cd BOSSA && \
    make && \
    cp bin/bossac /usr/local/bin

RUN pip3 install git+https://github.com/kendryte/kflash.py.git

ENV GO_RELEASE=1.17
RUN wget https://dl.google.com/go/go${GO_RELEASE}.linux-amd64.tar.gz && \
    tar xfv go${GO_RELEASE}.linux-amd64.tar.gz -C /usr/local && \
    rm go${GO_RELEASE}.linux-amd64.tar.gz
ENV PATH=${PATH}:/usr/local/go/bin

FROM tinygohci-base AS tinygohci-build
ENV TINYGO_RELEASE=0.19.0
ARG TINYGO_DOWNLOAD_URL=https://github.com/tinygo-org/tinygo/releases/download/v${TINYGO_RELEASE}/tinygo${TINYGO_RELEASE}.linux-amd64.tar.gz
RUN wget ${TINYGO_DOWNLOAD_URL} -O tinygo.tar.gz && \
    tar -xzf tinygo.tar.gz -C /usr/local && \
    rm tinygo.tar.gz
ENV PATH=${PATH}:/usr/local/tinygo/bin
RUN go get -d tinygo.org/x/drivers

RUN apt-get remove -y wget && \
    apt-get autoremove -y && \
    apt-get clean

CMD ["tinygo"]
