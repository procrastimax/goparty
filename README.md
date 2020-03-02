# goparty - the youtube music queue software

## What is this

This is a program for handling music at parties or just with some friends.
It allows downloading songs from youtube and playing them in an evenly distributed song queue.
All downloaded songs get saved as mp3 files. When trying to download the same song again, the program recognizes this and just queues the already downloaded song.
Also it is possible to see all mp3 files in a directory, so f.e. you can provide some of your favorite albums in a directory the program respects and then loads also these offline songs in.
All this works by creating a webserver on the admin machine and all smartphones/ laptop capable of using a webbrowser can access the song queue.
It is also possible for users to vote their favorite songs up, then they get reranked in the song queue.

## Why not use a normal playlist

The big advantage of this program is the ever expanding list of songs and the garanty for every user to hear his/her favorite song. Because the order of the queue can only be changed by the admin and by upvoting songs. On this way every user has the change to listen to his favorite song.
Also by downloading all songs only once (and not streaming them) the program could soon be used completely offline.

## What do I need for running this

You need a running [go setup](<https://golang.org/doc/install>).
Then go into this directory and execute `go build` or `go install`.

Also you need youtube-dl for downloading songs from youtube and ffmpeg to convert the downloaded song to mp3 files.

- [youtube-dl](<https://ytdl-org.github.io/youtube-dl>)
- [ffmpeg](<https://ffmpeg.org/>)

When using linux, your distro should provide those packages, so try downloading it with `apt-get install youtube-dl` (ubuntu), `xbps-install -Su youtube-dl` (void linux) ...

For windows users please make sure that that youtube-dl and ffmpeg get installed into your binary path directory. [Install tutorial for ffmpeg](<https://windowsloop.com/install-ffmpeg-windows-10/>)

Please keep youtube-dl always up-to-date. Because when using an older version of youtube-dl functionality cannot be guaranteed!

## Config

After the first start of the program a config.json file is generated. On linux systems you can find the config at ~/.config/goparty/config.json.
There you can *currently* set 4 parameters.

- **'downloadPath'** - sets the path in which the songs should get downloaded (THIS MUST BE SET)
- **'musicPath'** - sets the path from which offline music should get included (THIS MUST BE SET)
- 'upvotesForRerank' - sets the number of upvotes needed for a song to consider a reranking in the song queue
- 'allUserAdmin' - gives all users admin privileges for pausing the music or skipping a song

For best functionality please set the downloadPath inside the musicPath, so it looks like following:
`"downloadPath"="/home/procrastimax/music/goparty/"`
`"musicPath"="/home/procrastimax/music/"`
**You need to set a slash at the end of every path!**

Windows user need to escape all their slashes (`\\`)!
For example: 
`"downloadPath"="C:\\user\\music\\goparty\\"`
`"musicPath"="C:\\user\\music\\""`

## Screenshots

![Admin Page](screenshots/admin_page.png "Admin Page")

-------------------------------------------------------

![User Page](screenshots/user_view.png "User Page")

-------------------------------------------------------

![Offline Songs](screenshots/offline_song_page.png "Offline Song Page")
