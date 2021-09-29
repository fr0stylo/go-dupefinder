package filehash

import (
	"bufio"
	"crypto/sha256"
	"hash"
	"io"
	"log"
	"os"
)

type hashfile struct {
	h hash.Hash
}

type Hash interface {
	Hash(path string) ([]byte, error)
}

func New(hash hash.Hash) Hash {
	h := hash
	if h == nil {
		h = sha256.New()
	}

	return &hashfile{
		h: h,
	}
}

func (c *hashfile) Hash(path string) ([]byte, error) {
	fd, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer fd.Close()

	reader := bufio.NewReader(fd)
	buf := make([]byte, 1024)
	for {
		n, err := reader.Read(buf)
		if err != nil && err != io.EOF {
			log.Fatal(err)
		}

		buf := buf[:n]
		c.h.Write(buf)

		if err == io.EOF {
			break
		}
	}

	return c.h.Sum(nil), nil
}
