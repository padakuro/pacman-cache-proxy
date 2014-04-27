package main

import (
	"errors"
	"os"
)

type PkgFile struct {
	store *PkgStore
	repo  string
	arch  string
	fname string
	f     *os.File
}

func (pkg *PkgFile) Write(b []byte) (nr int, err error) {
	if pkg.f == nil {
		path := pkg.store.GetPkgPath(pkg)
		pkg.f, err = os.Create(path)
		if err != nil {
			return 0, err
		}
	}
	return pkg.f.Write(b)
}

func (pkg *PkgFile) Close() error {
	if pkg.f == nil {
		return errors.New("PkgFile was never written into")
	}
	return pkg.f.Close()
}
