FROM debian:stable-slim AS tinygohci-base

RUN apt-get clean && apt-get update && \
    apt-get install -y wget gcc avrdude git make build-essential libreadline-dev libwxgtk3.0-* python3 python3-pip openocd

# to address https://github.com/golang/go/issues/51253
COPY tools/docker/gitconfig.txt /etc/gitconfig

RUN git clone https://github.com/shumatech/BOSSA.git && \
    cd BOSSA && \
    make bossac && \
    cp bin/bossac /usr/local/bin

RUN pip3 install git+https://github.com/kendryte/kflash.py.git --break-system-packages

ENV GO_RELEASE=1.25.7
RUN wget https://dl.google.com/go/go${GO_RELEASE}.linux-amd64.tar.gz && \
    tar xfv go${GO_RELEASE}.linux-amd64.tar.gz -C /usr/local && \
    rm go${GO_RELEASE}.linux-amd64.tar.gz
ENV PATH=${PATH}:/usr/local/go/bin

FROM tinygohci-base AS tinygohci-build
ENV TINYGO_RELEASE=0.39.0
ARG TINYGO_DOWNLOAD_SHA=1234

COPY tools/docker/versions/${TINYGO_DOWNLOAD_SHA}.deb /tmp/tinygo.deb
RUN dpkg -i /tmp/tinygo.deb && rm /tmp/tinygo.deb
ENV PATH=${PATH}:/usr/local/tinygo/bin

RUN apt-get remove -y wget && \
    apt-get autoremove -y && \
    apt-get clean

CMD ["tinygo"]
