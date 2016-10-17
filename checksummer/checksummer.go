package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

func main() {
	wd, _ := os.Getwd()

	max := runtime.GOMAXPROCS(runtime.NumCPU())
	wg := &sync.WaitGroup{}
	ch := make(chan *os.File, max)
	for i := 0; i < max; i++ {
		wg.Add(1)
		go func() {
			for f := range ch {
				h := sha256.New()
				if _, err := io.Copy(h, f); err != nil {
					f.Close()
					log.Fatal(err)
				}

				fmt.Println(hex.EncodeToString(h.Sum(nil)))
				f.Close()
			}
			wg.Done()
		}()
	}

	err := filepath.Walk(wd, func(path string, fi os.FileInfo, err error) error {
		if err != nil && err != filepath.SkipDir {
			return err
		}
		if fi.IsDir() {
			if fi.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}

		if fi.Mode()&os.ModeSymlink == os.ModeSymlink {
			return nil
		}

		var f *os.File
		f, err = os.Open(path)
		if err != nil {
			return err
		}
		ch <- f

		return nil
	})

	close(ch)
	wg.Wait()
	if err != nil {
		log.Fatal(err)
	}
}
