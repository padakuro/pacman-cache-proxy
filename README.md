# pacman-cache-proxy

Very simple and stupid [pacman](https://wiki.archlinux.org/index.php/pacman) package cache written in Go. It is mostly based on the [httpdump example](https://github.com/elazarl/goproxy/blob/master/examples/httpdump/httpdump.go) from the goproxy library.

Beware of bugs and shitty code, it was programmed "to just work" :) If there's interest in this (or I use it more extensively than just for development containers), this will probably change.

## Why

When building many Docker containers, it is useful to have a local cache of the package files. Debian has apt-cacher-ng (and others), but somehow ArchLinux is missing something alike (closest would be [pacserve](http://xyne.archlinux.ca/projects/pacserve/)). Also an oppurtunity to check out [Go](http://golang.org/) (yeah, this is my first golang program :P).

## How

Pacman uses wget/curl and/or other tools to fetch packages from HTTP repositories. Most of these tools can use a proxy that is specified through environment variables. In this case, the `http_proxy` env variable.

The official repositories (core, extra, community) have the following URL structure:

```
/$repo/os/$arch/$packageFileName

# eg.
/core/os/x86_64/which-2.20-6-x86_64.pkg.tar.xz
```

This is used to extract the package file name which is unique within all those repositories. If the proxy sees such an URL, it will extract the information from above and download the package file to it's local cache directory and transparently serves the file to the requesting client.
The mirror used to actually download the file is the same as the one which was originally requested. Every subsequent request that matches the above package, regardless of mirror, will receive the file from the local cache directory.

Notes:

* The proxy does **NOT** cache the database files or any other files not matching the above URL structure.

## Build

```shell
git clone https://github.com/padakuro/pacman-cache-proxy.git
cd pacman-cache-proxy
go build
```

## Usage

### Run the proxy

```shell
# run it on 0.0.0.0:8080 (this is the default)
./pacman-cache-proxy

# run it on a specific ip and port
./pacman-cache-proxy -l 192.168.13.37:10042
```

* `-v`: enables verbose output
* `-l [address]:port`: listen address/port combination, address is optional (default to 0.0.0.0)
* `-p dir`: cache directory path, by default current working directory

### Use the proxy

```shell
# assuming the proxy runs on localhost
sudo env http_proxy="127.0.0.1:8080" pacman -Syu

# in a Dockerfile
# RUN env http_proxy="192.168.13.37:10042" pacman --noconfirm --needed -S which
```

## License

MIT
