package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func main() {
	for {
		// Ввод длинного URL от пользователя
		endpoint := "http://localhost:8080/"

		// приглашение в консоли
		fmt.Println("==========================================")
		fmt.Println("🌐 Добро пожаловать в сокращатель URL")
		fmt.Println("==========================================")
		fmt.Println("Выберите одну из опций:")
		fmt.Println("1️⃣  - Использовать валидный дефолтный URL: http://example.com")
		fmt.Println("2️⃣  - Использовать невалидный дефолтный URL: http://invalid-url")
		fmt.Println("3️⃣  - Ввести свой URL")
		fmt.Println("4️⃣  - Выход")

		// Считываем выбор пользователя
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan() // считываем выбор
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

		// добавляем HTTP-клиент
		client := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse // запретить автоматическую обработку редиректов
			},
		}
		// пишем запрос
		req, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(longURL))
		if err != nil {
			fmt.Println("❌ Ошибка при создании запроса:", err)
			return
		}
		// Указываем заголовок Content-Type
		req.Header.Set("Content-Type", "text/plain")

		// Выполняем запрос и обрабатываем ответ
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("❌ Ошибка выполнения запроса:", err)
			return
		}
		defer resp.Body.Close()

		// Читаем короткий URL из ответа
		responseBody, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("❌ Ошибка чтения ответа:", err)
			return
		}

		if resp.StatusCode != http.StatusCreated {
			fmt.Printf("❌ Ошибка сокращения URL:\n%s\n%s\n", resp.Status, string(responseBody))
			return
		}

		shortURLStr := string(responseBody)
		fmt.Println("✅ Короткий URL: " + shortURLStr)

		// Предлагаем пользователю либо нажать Enter для использования этого URL, либо ввести свой
		fmt.Print("\n🔀 Нажмите Enter, чтобы использовать этот короткий URL, или введите свой: ")
		scanner.Scan()
		userInput := strings.TrimSpace(scanner.Text())

		var id string
		if userInput == "" {
			// Если нажата Enter, используем уже сгенерированный короткий URL
			parts := strings.Split(shortURLStr, "/")
			id = parts[len(parts)-1]
			fmt.Println("✔️ Вы используете короткий URL:", shortURLStr)
		} else {
			// Если введен пользовательский URL, берем его идентификатор
			parts := strings.Split(userInput, "/")
			id = parts[len(parts)-1]
			fmt.Println("✔️ Вы ввели свой URL с идентификатором:", id)
		}
		// Используем идентификатор для перенаправления
		redirectURL := "http://localhost:8080/" + id

		fmt.Println("\n⏳ Пытаемся перенаправить по короткому URL...")

		getReq, err := http.NewRequest(http.MethodGet, redirectURL, nil)
		if err != nil {
			fmt.Println("❌ Ошибка при создании запроса:", err)
			return
		}
		// Указываем заголовок Content-Type
		getReq.Header.Set("Content-Type", "text/plain")

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
		scanner.Scan()
		again := strings.TrimSpace(scanner.Text())
		if strings.ToLower(again) != "да" {
			fmt.Println("👋 До свидания!")
			break
		}
	}
}
