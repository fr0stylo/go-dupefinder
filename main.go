package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/fr0stylo/go-dupfinder/filehash"
)

var db map[string][]string
var wg sync.WaitGroup

func walkFunc(pathChan chan string) func(path string, info fs.FileInfo, err error) error {
	return func(path string, info fs.FileInfo, err error) error {
		if !info.IsDir() {
			pathChan <- path
			wg.Add(1)
		}

		return nil
	}
}

func walkThroughFiles(rootDir string) chan string {
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
	Hash []byte
}

func walkFiles(pathC chan string, storeC chan *FileDef) {
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

		if _, ok := db[hash]; !ok {
			db[hash] = make([]string, 0)
		}
		val := db[hash]
		val = append(val, fp)
		db[hash] = val
		wg.Done()
	}
}

func init() {
	db = make(map[string][]string)
}

func main() {
	log.Print(os.Args)
	if len(os.Args) < 2 {
		log.Fatal("Not enough arguments")
	}

	log.Printf("Wroking on file path %s", os.Args[1])
	pathC := walkThroughFiles(os.Args[1])
	for i := 0; i < 10; i++ {
		go walkFiles(pathC, nil)
	}
	wg.Wait()

	for k, v := range db {
		if len(v) > 1 {
			log.Printf("%s ->\n", k)
			for _, fp := range v {
				log.Printf("  |- %s n", fp)
			}

		}
	}

	log.Print("Done !")
}
