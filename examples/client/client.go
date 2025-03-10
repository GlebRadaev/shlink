package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	port := flag.String("a", ":8080", "HTTP server address")
	flag.Parse()
	baseEndpoint := fmt.Sprintf("http://localhost%s/", *port)

	client := &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	scanner := bufio.NewScanner(os.Stdin)

	for {
		// приглашение в консоли
		fmt.Println("==========================================")
		fmt.Println("🌐 Добро пожаловать в сокращатель URL")
		fmt.Println("==========================================")
		fmt.Println("Выберите одну из опций:")
		fmt.Println("1️⃣  - Использовать валидный дефолтный URL: http://example.com")
		fmt.Println("2️⃣  - Использовать невалидный дефолтный URL: http://invalid-url")
		fmt.Println("3️⃣  - Ввести свой URL")
		fmt.Println("4️⃣  - Выход")

		if !scanner.Scan() {
			fmt.Println("❌ Ошибка чтения ввода")
			return
		}
		option := strings.TrimSpace(scanner.Text())

		if option == "4" {
			fmt.Println("👋 До свидания!")
			break
		}

		var longURL string

		switch option {
		case "1":
			longURL = "http://example.com" // валидный URL по умолчанию
			fmt.Println("🔗 Вы выбрали валидный URL:", longURL)
		case "2":
			longURL = "http://invalid-url" // невалидный URL по умолчанию
			fmt.Println("🚫 Вы выбрали невалидный URL:", longURL)
		case "3":
			fmt.Print("Введите длинный URL: ")
			scanner.Scan() // считываем URL
			longURL = strings.TrimSpace(scanner.Text())
			fmt.Println("🔗 Вы ввели URL:", longURL)
		default:
			fmt.Println("❌ Некорректный выбор. Попробуйте снова.")
			return
		}
		fmt.Println("\n⏳ Пытаемся сократить ваш URL...")

		// пишем запрос
		req, err := http.NewRequest(http.MethodPost, baseEndpoint, strings.NewReader(longURL))
		if err != nil {
			fmt.Println("❌ Ошибка при создании запроса:", err)
			return
		}
		// Указываем заголовок Content-Type
		req.Header.Add("Content-Type", "text/plain")

		// Выполняем запрос и обрабатываем ответ
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("❌ Ошибка выполнения запроса:", err.Error())
			return
		}

		// Читаем короткий URL из ответа
		responseBody, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("❌ Ошибка чтения ответа:", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			fmt.Printf("❌ Ошибка сокращения URL:\n%s\n%s\n", resp.Status, string(responseBody))
			return
		}

		shortURLStr := string(responseBody)
		fmt.Println("✅ Короткий URL: " + shortURLStr)

		// Предлагаем пользователю либо нажать Enter для использования этого URL, либо ввести свой
		fmt.Print("\n🔀 Нажмите Enter, чтобы использовать этот короткий URL, или введите свой: ")
		if !scanner.Scan() {
			fmt.Println("❌ Ошибка чтения ввода")
			return
		}
		userInput := strings.TrimSpace(scanner.Text())

		var id string
		if userInput == "" {
			// Если нажата Enter, используем уже сгенерированный короткий URL
			parts := strings.Split(shortURLStr, "/")
			id = parts[len(parts)-1]
			fmt.Println("✔️ Вы используете короткий URL:", id)
		} else {
			// Если введен пользовательский URL, берем его идентификатор
			parts := strings.Split(userInput, "/")
			id = parts[len(parts)-1]
			fmt.Println("✔️ Вы ввели свой URL с идентификатором:", id)
		}

		fmt.Println("\n⏳ Пытаемся перенаправить по короткому URL...")
		getReq, err := http.NewRequest(http.MethodGet, baseEndpoint+id, nil)
		if err != nil {
			fmt.Println("❌ Ошибка при создании запроса:", err)
			return
		}
		// Указываем заголовок Content-Type
		getReq.Header.Add("Content-Type", "text/plain")

		// Выполняем запрос и обрабатываем ответ
		respGet, err := client.Do(getReq)
		if err != nil {
			fmt.Println("❌ Ошибка выполнения запроса:", err)
			return
		}
		defer respGet.Body.Close()

		// Проверяем код состояния ответа
		if respGet.StatusCode == http.StatusTemporaryRedirect {
			// Если был редирект, получаем оригинальный URL из заголовка "Location"
			originalURL := respGet.Header.Get("Location")
			if originalURL != "" {
				fmt.Println("✅ Исходный URL:", originalURL)
			} else {
				fmt.Println("❌ Заголовок 'Location' отсутствует в ответе")
			}
		} else {
			fmt.Println("❌ Ошибка перенаправления:", respGet.Status)
		}

		fmt.Println("\n🎉 Завершено! Хотите повторить? (да/нет): ")
		if !scanner.Scan() {
			fmt.Println("❌ Ошибка чтения ввода")
			return
		}
		again := strings.TrimSpace(scanner.Text())
		if strings.ToLower(again) != "да" {
			fmt.Println("👋 До свидания!")
			break
		}
	}
}
