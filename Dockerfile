FROM golang:latest AS builder

WORKDIR /go/src
COPY . .

RUN cp others/lib/libpv_porcupine.so /usr/lib/libpv_porcupine.so
RUN cp others/include/picovoice.h /usr/include/picovoice.h
RUN cp others/include/pv_porcupine.h /usr/include/pv_porcupine.h

RUN go build

RUN apt-get update -qq && apt-get install -y -q --no-install-recommends xz-utils
RUN wget https://youtube-dl.org/downloads/latest/youtube-dl
RUN chmod a+rx youtube-dl
RUN wget https://johnvansickle.com/ffmpeg/releases/ffmpeg-release-amd64-static.tar.xz
RUN tar xf ffmpeg-release-amd64-static.tar.xz
RUN mv ffmpeg-*-amd64-static/ffmpeg ffmpeg

FROM golang:latest AS prod
RUN mkdir config
COPY --from=builder /go/src/others/lib/libpv_porcupine.so /usr/lib/libpv_porcupine.so
COPY --from=builder /go/src/others /go/others
COPY --from=builder /go/src/senko /go/bin/senko
COPY --from=builder /go/src/youtube-dl /usr/bin/youtube-dl
COPY --from=builder /go/src/ffmpeg /usr/bin/ffmpeg
CMD ["/go/bin/senko"]