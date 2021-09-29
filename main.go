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

var db map[string][]*FileDef
var wg sync.WaitGroup
var mux sync.Mutex

func walkFunc(pathChan chan *FileDef) func(path string, info fs.FileInfo, err error) error {
	return func(path string, info fs.FileInfo, err error) error {
		if !info.IsDir() {
			pathChan <- &FileDef{
				Path: path,
				Size: info.Size(),
			}
			wg.Add(1)
		}

		return nil
	}
}

func walkThroughFilesRoutine(rootDir string) chan *FileDef {
	pathC := make(chan *FileDef, 100)

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
	Size int64
}

func hashFilesRoutine(pathC chan *FileDef, storeC chan *FileDef) {
	run := true
	for run {
		fp, ok := <-pathC
		run = ok

		fh := filehash.New(nil)
		h, err := fh.Hash(fp.Path)
		if err != nil {
			log.Print(err)
			wg.Done()
			continue
		}

		hash := fmt.Sprintf("%x", h)
		fp.Hash = hash
		storeC <- fp
	}
}

func storeToDbRoutine() chan *FileDef {
	storeC := make(chan *FileDef)

	go func() {
		for {
			fd := <-storeC
			mux.Lock()
			if _, ok := db[fd.Hash]; !ok {
				db[fd.Hash] = make([]*FileDef, 0)
			}
			val := db[fd.Hash]
			val = append(val, fd)
			db[fd.Hash] = val
			mux.Unlock()

			wg.Done()
		}
	}()

	return storeC
}

func init() {
	db = make(map[string][]*FileDef)
}

func main() {
	parralel := flag.Int("p", 10, "sets paralelization level for hashing")
	sizeThreshold := flag.Int("st", 0, "sets size threshold in kb")
	help := flag.Bool("h", false, "help command")
	flag.Parse()
	rootPath := flag.Arg(0)

	if *help {
		flag.PrintDefaults()
		os.Exit(0)
	}

	log.Printf("Wroking on file path %s", rootPath)
	pathC := walkThroughFilesRoutine(rootPath)
	storeC := storeToDbRoutine()

	for i := 0; i < *parralel; i++ {
		go hashFilesRoutine(pathC, storeC)
	}
	wg.Wait()

	for k, v := range db {
		if len(v) > 1 {
			size := float64(v[0].Size) / 1024
			if size > float64(*sizeThreshold) {
				log.Printf("%s ->\n", k)
				log.Printf("Single file size %f kb \n", size)
				log.Printf("All duplicates takes %f kb \n", float64(v[0].Size*int64(len(v)))/1024)

				for _, fp := range v {
					log.Printf("  |- %s", fp.Path)
				}

				log.Print()
			}
		}
	}

	log.Print("Done !")
}
