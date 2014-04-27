package main

import (
	"github.com/elazarl/goproxy"
	"io"
	"log"
	"net/http"
	"os"
	"path"
)

func NewTeeReadCloser(r io.ReadCloser, w io.WriteCloser) io.ReadCloser {
	return &TeeReadCloser{io.TeeReader(r, w), w, r}
}

type PkgStore struct {
	path  string
	errch chan error
}

func (store *PkgStore) NewPkg(repo string, arch string, fname string) *PkgFile {
	return &PkgFile{store, repo, arch, fname, nil}
}

func (store *PkgStore) GetPkgDir(pkg *PkgFile) string {
	return path.Join(store.path, pkg.repo, "os", pkg.arch)
}

func (store *PkgStore) GetPkgPath(pkg *PkgFile) string {
	return path.Join(store.GetPkgDir(pkg), pkg.fname)
}

func (store *PkgStore) HasPkg(pkg *PkgFile) bool {
	path := store.GetPkgPath(pkg)
	_, err := os.Stat(path)
	return err == nil
}

func (store *PkgStore) PutPkg(resp *http.Response, ctx *goproxy.ProxyCtx) {
	pkg := ctx.UserData.(*PkgFile)
	if resp != nil {
		log.Printf("Storing package: %s/%s/%s", pkg.repo, pkg.arch, pkg.fname)

		dir := store.GetPkgDir(pkg)
		if _, err := os.Stat(dir); err != nil {
			err := os.MkdirAll(dir, 0755)
			if err != nil {
				log.Printf("Could not store package: could not create directory, " + dir)
				return
			}
		}

		resp.Body = NewTeeReadCloser(resp.Body, pkg)
	}
}

func (store *PkgStore) Close() error {
	return <-store.errch
}

type TeeReadCloser struct {
	r io.Reader
	w io.WriteCloser
	c io.Closer
}

func (t *TeeReadCloser) Read(b []byte) (int, error) {
	return t.r.Read(b)
}

func (t *TeeReadCloser) Close() error {
	err1 := t.c.Close()
	err2 := t.w.Close()
	if err1 == nil && err2 == nil {
		return nil
	}
	if err1 != nil {
		return err2
	}
	return err1
}
