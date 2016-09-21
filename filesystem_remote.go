package main

import (
	"io/ioutil"
	"os"
)

type filesystemRemote struct {
	dir string
}

func (fr *filesystemRemote) getObjects(hashes []string) (contents [][]byte, err error) {
	for _, hash := range hashes {
		b, err := ioutil.ReadFile(fr.dir + "/.cap/objects/" + hash)
		if err != nil {
			return nil, err
		}
		contents = append(contents, b)
	}
	return contents, nil
}

func (fr *filesystemRemote) listObjects() (hashes []string, err error) {
	dir, err := os.Open(fr.dir + "/.cap/objects")
	if err != nil {
		return nil, err
	}
	defer dir.Close()
	return dir.Readdirnames(0)
}

func (fr *filesystemRemote) listBranches() (refs [][2]string, err error) {
	files, err := ioutil.ReadDir(fr.dir + "/.cap/refs/heads")
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		b, err := ioutil.ReadFile(fr.dir + "/.cap/refs/heads/" + file.Name())
		if err != nil {
			return nil, err
		}
		refs = append(refs, [2]string{file.Name(), string(b)})
	}
	return refs, nil
}
