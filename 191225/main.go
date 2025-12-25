package main

import (
	"fmt"
	"time"
)

func newsFeed(newsChan chan<- string, stopChan <-chan bool) {
	messages := []string{
		"Цены на нефть упали",
		"Новый курс по программированию",
		"Выборы президента начались",
		"Открыта новая станция метро",
		"Курс доллара обновил максимум",
	}

	for _, msg := range messages {
		select {
		case newsChan <- msg:
			time.Sleep(1 * time.Second)
		case <-stopChan:
			fmt.Println("newsFeed: получен сигнал остановки")
			return
		}
	}
	close(newsChan)
	fmt.Println("newsFeed: все сообщения отправлены")
}
func socialMedia(socialChan chan<- string, stopChan <-chan bool) {
	messages := []string{
		"Иван выложил новое фото",
		"Мария отметила вас на фото",
		"Новый тренд в TikTok",
		"Обновление Instagram",
		"ИИ создал новую музыку",
	}

	for _, msg := range messages {
		select {
		case socialChan <- msg:
			time.Sleep(700 * time.Millisecond)
		case <-stopChan:
			fmt.Println("socialMedia: получен сигнал остановки")
			return
		}
	}
	close(socialChan)
	fmt.Println("socialMedia: все сообщения отправлены")
}

func main() {

	newsChan := make(chan string)
	socialChan := make(chan string)
	stopChan := make(chan bool)

	go func() {
		time.Sleep(5 * time.Second)
		fmt.Println("\nGraceful Shutdown: инициируем остановку...")
		close(stopChan)
	}()

	go newsFeed(newsChan, stopChan)
	go socialMedia(socialChan, stopChan)

	fmt.Println("=== Запуск обработчика сообщений ===")
	fmt.Println("Программа завершится через 5 секунд...")

MainLoop:
	for {
		select {

		case news, ok := <-newsChan:
			if ok {
				fmt.Printf("Новость: %s\n", news)
			} else {
				fmt.Println("Канал новостей закрыт")
				newsChan = nil
			}

		case social, ok := <-socialChan:
			if ok {
				fmt.Printf("Соцсети: %s\n", social)
			} else {
				fmt.Println("Канал соцсетей закрыт")
				socialChan = nil
			}

		case <-stopChan:
			fmt.Println("\nПолучен сигнал остановки, завершаем работу...")

			time.Sleep(1 * time.Second)
			break MainLoop

		default:
			if newsChan == nil && socialChan == nil {
				fmt.Println("\nВсе каналы закрыты, завершаем работу")
				break MainLoop
			}
			fmt.Print(".")
			time.Sleep(200 * time.Millisecond)
		}
	}

	fmt.Println("Ожидаем завершения горутин...")
	time.Sleep(500 * time.Millisecond)
	fmt.Println("=== Программа завершена корректно ===")
}
