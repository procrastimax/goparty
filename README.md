# yt-queue - the youtube music queue software

##TODO
- user restriction (X songs by a user cant be played in a row -> shuffle the download/playing queue)
- a user should not add more than X songs per hour/day
- user voting for bringing up downloaded songs which are currently in a queue

## Uses libraries & Binaries
- beep library -> https://github.com/faiface/beep
- oto (low level sound handling) -> https://github.com/hajimehoshi/oto
- go-mp3 (mp3) -> https://github.com/hajimehoshi/go-mp3
- youtube-dl -> https://ytdl-org.github.io/youtube-dl
- ffmpeg -> requiered by youtube-dl for converting to mp3
- alsa-lib-devel
## Requirements
- libasound2-dev
- youtube-dl (https://ytdl-org.github.io/youtube-dl/download.html)
- pulseaudio mustn't be installed!
