package main

import (
	"bufio"
	"flag"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"regexp"

	"github.com/elazarl/goproxy"
	"github.com/elazarl/goproxy/transport"
)

func NewPkgStore(basepath string) (*PkgStore, error) {
	err := os.MkdirAll(basepath, 0755)
	if err != nil {
		return nil, err
	}
	store := &PkgStore{basepath, make(chan error)}
	return store, nil
}

func NewCachedResponse(r *http.Request, store *PkgStore, pkg *PkgFile) *http.Response {
	path := store.GetPkgPath(pkg)

	info, _ := os.Stat(path)

	resp := &http.Response{}
	resp.Request = r
	resp.TransferEncoding = r.TransferEncoding
	resp.Header = make(http.Header)
	resp.Header.Add("Content-Type", "application/octet-stream")
	resp.StatusCode = 200

	file, _ := os.Open(path)
	buf := bufio.NewReader(file)
	resp.ContentLength = info.Size()
	resp.Body = ioutil.NopCloser(buf)

	return resp
}

func main() {
	verbose := flag.Bool("v", false, "should every proxy request be logged to stdout")
	addr := flag.String("l", ":8080", "on which address should the proxy listen")
	flag.Parse()

	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = *verbose

	store, err := NewPkgStore("/tmp/packages")
	if err != nil {
		log.Fatal("Could not create package cache directory")
	}

	tr := transport.Transport{Proxy: transport.ProxyFromEnvironment}
	r := regexp.MustCompile(`/([^/]+)/os/(i686|x86_64)/(.+\.pkg\.tar.+)`)

	proxy.OnRequest().DoFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		uri := req.URL.RequestURI()
		if r.MatchString(uri) == true {
			pkgInfo := r.FindStringSubmatch(uri)
			pkgFile := store.NewPkg(pkgInfo[1], pkgInfo[2], pkgInfo[3])
			log.Printf("Request: %s/%s/%s", pkgFile.repo, pkgFile.arch, pkgFile.fname)

			if store.HasPkg(pkgFile) {
				log.Printf("Serving cached package: %s/%s/%s", pkgFile.repo, pkgFile.arch, pkgFile.fname)
				return req, NewCachedResponse(req, store, pkgFile)
			} else {
				ctx.RoundTripper = goproxy.RoundTripperFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (resp *http.Response, err error) {
					resp, err = tr.RoundTrip(req)
					ctx.UserData = pkgFile
					return
				})
			}
		}
		return req, nil
	})
	proxy.OnResponse().DoFunc(func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		if ctx.UserData != nil {
			store.PutPkg(resp, ctx)
		}
		return resp
	})

	l, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatal("Listen: ", err)
	}

	sl := newStoppableListener(l)
	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt)
	go func() {
		<-ch
		log.Println("Got SIGINT exiting")
		sl.Add(1)
		sl.Close()
		sl.Done()
	}()

	log.Println("Starting Proxy")
	http.Serve(sl, proxy)
	sl.Wait()
	log.Println("All connections closed - exit")
}
