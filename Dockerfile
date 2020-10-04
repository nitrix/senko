FROM golang:latest AS builder

WORKDIR /opt

RUN git clone https://github.com/nitrix/porcupine porcupine

RUN apt-get update -qq && apt-get install -y -q --no-install-recommends xz-utils
RUN wget https://youtube-dl.org/downloads/latest/youtube-dl
RUN chmod a+rx youtube-dl
RUN wget https://johnvansickle.com/ffmpeg/releases/ffmpeg-release-amd64-static.tar.xz
RUN tar xf ffmpeg-release-amd64-static.tar.xz
RUN mv ffmpeg-*-amd64-static/ffmpeg ffmpeg

COPY . /opt/senko
WORKDIR /opt/senko

RUN go build -ldflags "-s -w"

FROM golang:latest AS release
COPY --from=builder /opt/youtube-dl /usr/bin/youtube-dl
COPY --from=builder /opt/ffmpeg /usr/bin/ffmpeg
COPY --from=builder /opt/senko/assets /opt/senko/assets
COPY --from=builder /opt/senko/senko /opt/senko/senko
COPY --from=builder /opt/porcupine/lib/libpv_porcupine.so /usr/lib/libpv_porcupine.so

WORKDIR /opt/senko
CMD ["/opt/senko/senko"]