package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

type fileData struct {
	path string
	size int64
	name string
}

func main() {

	dirPtr := flag.String("dir", ".", "папка для поиска")
	workersPtr := flag.Int("workers", 5, "сколько горутин")
	flag.Parse()

	dir := *dirPtr
	workers := *workersPtr

	if _, err := os.Stat(dir); err != nil {
		fmt.Printf("Ошибка: папка '%s' не найдена\n", dir)
		return
	}

	fmt.Printf("Ищем дубликаты в: %s\n", dir)
	fmt.Printf("Рабочих горутин: %d\n\n", workers)

	filesChan := make(chan string, 1000)
	resultsChan := make(chan fileData, 1000)
	var wg sync.WaitGroup

	var processed int64 = 0

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for filePath := range filesChan {

				info, err := os.Stat(filePath)
				if err != nil {
					continue
				}

				if info.IsDir() {
					continue
				}

				resultsChan <- fileData{
					path: filePath,
					size: info.Size(),
					name: info.Name(),
				}

				atomic.AddInt64(&processed, 1)
			}
		}(i)
	}

	allFiles := make([]fileData, 0)
	var mu sync.Mutex
	var collectorWg sync.WaitGroup
	collectorWg.Add(1)
	go func() {
		defer collectorWg.Done()
		for file := range resultsChan {
			mu.Lock()
			allFiles = append(allFiles, file)
			mu.Unlock()
		}
	}()

	stopProgress := make(chan bool)
	go func() {
		ticker := time.NewTicker(300 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				count := atomic.LoadInt64(&processed)
				fmt.Printf("\rНайдено файлов: %d", count)
			case <-stopProgress:
				count := atomic.LoadInt64(&processed)
				fmt.Printf("\rНайдено файлов: %d\n", count)
				return
			}
		}
	}()

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if !info.IsDir() {
			filesChan <- path
		}
		return nil
	})

	close(filesChan)
	wg.Wait()
	close(resultsChan)
	collectorWg.Wait()

	stopProgress <- true
	time.Sleep(50 * time.Millisecond)

	if err != nil {
		fmt.Printf("Ошибка при обходе: %v\n", err)
	}

	fmt.Printf("Всего файлов: %d\n\n", len(allFiles))

	groups := make(map[string][]fileData)

	for _, file := range allFiles {
		key := fmt.Sprintf("%s|%d", file.name, file.size)
		groups[key] = append(groups[key], file)
	}

	foundAny := false
	for _, files := range groups {
		if len(files) > 1 {
			foundAny = true

			fmt.Printf("Дубликаты (%d файлов, имя: %s, размер: %d байт):\n",
				len(files), files[0].name, files[0].size)

			for i, f := range files {
				fmt.Printf("  %d. %s\n", i+1, f.path)
			}
			fmt.Println()
		}
	}

	if !foundAny {
		fmt.Println("Дубликатов не найдено!")
		return
	}

	duplicateCount := 0
	spaceWasted := int64(0)

	for _, files := range groups {
		if len(files) > 1 {
			duplicateCount += len(files) - 1
			spaceWasted += int64(len(files)-1) * files[0].size
		}
	}

	fmt.Println("=== Статистика ===")
	fmt.Printf("Лишних копий: %d\n", duplicateCount)

	if spaceWasted > 1024*1024 {
		fmt.Printf("Можно освободить: %.2f MB\n", float64(spaceWasted)/(1024*1024))
	} else if spaceWasted > 1024 {
		fmt.Printf("Можно освободить: %.2f KB\n", float64(spaceWasted)/1024)
	} else {
		fmt.Printf("Можно освободить: %d байт\n", spaceWasted)
	}
}
