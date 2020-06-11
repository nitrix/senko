FROM golang:latest AS builder

WORKDIR /go/src
COPY . .

RUN go build

RUN apt-get update -qq && apt-get install -y -q --no-install-recommends xz-utils
RUN wget https://youtube-dl.org/downloads/latest/youtube-dl
RUN chmod a+rx youtube-dl
RUN wget https://johnvansickle.com/ffmpeg/releases/ffmpeg-release-amd64-static.tar.xz
RUN tar xf ffmpeg-release-amd64-static.tar.xz
RUN mv ffmpeg-*-amd64-static/ffmpeg ffmpeg

FROM golang:latest AS prod
RUN mkdir config
COPY --from=builder /go/src/senko /go/bin/senko
COPY --from=builder /go/src/youtube-dl /usr/bin/youtube-dl
COPY --from=builder /go/src/ffmpeg /usr/bin/ffmpeg
CMD ["/go/bin/senko"]