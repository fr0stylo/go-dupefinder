package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/fr0stylo/go-dupefinder/filehash"
)

var db map[string][]string
var wg sync.WaitGroup
var mux sync.Mutex

func walkFunc(pathChan chan string) func(path string, info fs.FileInfo, err error) error {
	return func(path string, info fs.FileInfo, err error) error {
		if !info.IsDir() {
			pathChan <- path
			wg.Add(1)
		}

		return nil
	}
}

func walkThroughFilesRoutine(rootDir string) chan string {
	pathC := make(chan string, 100)

	wg.Add(1)
	go func() {
		filepath.Walk(rootDir, walkFunc(pathC))
		defer wg.Done()
	}()

	return pathC
}

type FileDef struct {
	Path string
	Hash string
}

func hashFilesRoutine(pathC chan string, storeC chan *FileDef) {
	run := true
	for run {
		fp, ok := <-pathC
		run = ok

		fh := filehash.New(nil)
		h, err := fh.Hash(fp)
		if err != nil {
			log.Print(err)
			wg.Done()
			continue
		}

		hash := fmt.Sprintf("%x", h)
		storeC <- &FileDef{Hash: hash, Path: fp}
	}
}

func storeToDbRoutine() chan *FileDef {
	storeC := make(chan *FileDef)

	go func() {
		for {
			fd := <-storeC
			mux.Lock()
			if _, ok := db[fd.Hash]; !ok {
				db[fd.Hash] = make([]string, 0)
			}
			val := db[fd.Hash]
			val = append(val, fd.Path)
			db[fd.Hash] = val
			mux.Unlock()

			wg.Done()
		}
	}()

	return storeC
}

func init() {
	db = make(map[string][]string)
}

func main() {
	parralel := flag.Int("p", 10, "sets paralelization level for hashing")
	log.Print(os.Args)
	if len(os.Args) < 2 {
		log.Fatal("Not enough arguments")
	}

	log.Printf("Wroking on file path %s", os.Args[1])
	pathC := walkThroughFilesRoutine(os.Args[1])
	storeC := storeToDbRoutine()

	for i := 0; i < *parralel; i++ {
		go hashFilesRoutine(pathC, storeC)
	}
	wg.Wait()

	for k, v := range db {
		if len(v) > 1 {
			log.Printf("%s ->\n", k)
			for _, fp := range v {
				log.Printf("  |- %s", fp)
			}

		}
	}

	log.Print("Done !")
}
