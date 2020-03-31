FROM golang:latest AS builder

WORKDIR /go/src/app
COPY . .

RUN go build

RUN apt-get update -qq && apt-get install -y -q --no-install-recommends xz-utils

RUN wget https://youtube-dl.org/downloads/latest/youtube-dl
RUN chmod a+rx youtube-dl
RUN mv youtube-dl /usr/bin/youtube-dl

RUN wget https://johnvansickle.com/ffmpeg/releases/ffmpeg-release-amd64-static.tar.xz
RUN tar xf ffmpeg-release-amd64-static.tar.xz
RUN mv ffmpeg-4.2.2-amd64-static/ffmpeg ffmpeg
RUN mv ffmpeg /usr/bin/ffmpeg
RUN rm -rf ffmpeg-4.2.2-amd64-static

RUN mv senko /bin/senko

CMD ["/bin/senko"]