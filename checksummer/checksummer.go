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
	"strings"
	"sync"
)

func main() {
	wd, _ := os.Getwd()

	max := runtime.GOMAXPROCS(runtime.NumCPU())
	wg := &sync.WaitGroup{}
	ch := make(chan string, max)
	for i := 0; i < max; i++ {
		wg.Add(1)
		go func() {
			for path := range ch {
				f, err := os.Open(path)
				if err != nil {
					continue
				}

				h := sha256.New()
				if _, err = io.Copy(h, f); err != nil {
					f.Close()
					log.Fatal(err)
				}

				fmt.Println(strings.TrimPrefix(path, wd), hex.EncodeToString(h.Sum(nil)))
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

		ch <- path
		//var f *os.File
		//f, err = os.Open(path)
		//if err != nil {
		//return err
		//}
		//defer f.Close()

		//h := sha256.New()
		//if _, err = io.Copy(h, f); err != nil {
		//log.Fatal(err)
		//}

		//fmt.Println(strings.TrimPrefix(path, wd), hex.EncodeToString(h.Sum(nil)))
		return nil
	})

	close(ch)
	wg.Wait()
	if err != nil {
		log.Fatal(err)
	}
}
