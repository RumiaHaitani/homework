package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
)

type FileProcessor interface {
	Process(path string, entry fs.DirEntry) (int64, error)
}

type SizeCounter struct{}

func (sc SizeCounter) Process(path string, entry fs.DirEntry) (int64, error) {
	info, err := entry.Info()
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

var sem = make(chan struct{}, 20)

func walkConcurrent(path string, processor FileProcessor, sizes chan<- int64, wg *sync.WaitGroup) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return
	}

	for _, entry := range entries {
		fullPath := filepath.Join(path, entry.Name())

		if entry.IsDir() {
			walkConcurrent(fullPath, processor, sizes, wg)
		} else {
			wg.Add(1)
			go func(p string, e fs.DirEntry) {
				defer wg.Done()

				sem <- struct{}{}
				defer func() { <-sem }()

				size, _ := processor.Process(p, e)
				sizes <- size
			}(fullPath, entry)
		}
	}
}

func main() {
	path := "."
	processor := SizeCounter{}

	sizes := make(chan int64)
	var wg sync.WaitGroup

	go func() {
		walkConcurrent(path, processor, sizes, &wg)
		wg.Wait()
		close(sizes)
	}()

	var totalSize int64
	for size := range sizes {
		totalSize += size
	}

	fmt.Printf("Total size: %.2f MB\n", float64(totalSize)/1024/1024)
}
