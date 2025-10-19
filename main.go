package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
)

// Структура для декодирования запроса от нашего фронтенда
type FrontendRequest struct {
	Prompt string `json:"prompt"`
}

// Структура для отправки запроса на внешний API
type OpenAIRequest struct {
	Messages []Message `json:"messages"`
}

// Структура для ответа от внешнего API (упрощенная)
type OpenAIResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// URL целевого API
const apiURL = "https://openai-hub.neuraldeep.tech/v1/chat/completions"

// handleChat - обработчик для /api/chat
func handleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	// 1. Читаем запрос от нашего фронтенда
	var req FrontendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Error decoding request: %v", err), http.StatusBadRequest)
		return
	}
	if req.Prompt == "" {
		http.Error(w, "Prompt cannot be empty", http.StatusBadRequest)
		return
	}
	log.Printf("Получен промпт: %s", req.Prompt)

	// 2. Получаем API ключ из переменных окружения (безопасный способ)
	apiKey := os.Getenv("LITELLM_API_KEY")
	if apiKey == "" {
		log.Println("Внимание: переменная окружения LITELLM_API_KEY не установлена.")
		http.Error(w, "API key is not configured on the server", http.StatusInternalServerError)
		return
	}

	// 3. Формируем тело запроса для внешнего API
	apiRequest := OpenAIRequest{
		Messages: []Message{
			{
				Role:    "user",
				Content: req.Prompt,
			},
		},
	}
	apiRequestBody, err := json.Marshal(apiRequest)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error marshalling API request: %v", err), http.StatusInternalServerError)
		return
	}

	// 4. Создаем и отправляем POST запрос на внешний API
	httpRequest, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(apiRequestBody))
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating HTTP request: %v", err), http.StatusInternalServerError)
		return
	}
	httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.Header.Set("x-litellm-api-key", apiKey)

	client := &http.Client{}
	httpResponse, err := client.Do(httpRequest)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error sending request to API: %v", err), http.StatusInternalServerError)
		return
	}
	defer httpResponse.Body.Close()

	// 5. Читаем ответ от внешнего API
	responseBody, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading API response: %v", err), http.StatusInternalServerError)
		return
	}

	if httpResponse.StatusCode != http.StatusOK {
		log.Printf("API вернул ошибку: %s", string(responseBody))
		http.Error(w, fmt.Sprintf("API returned non-200 status: %s", string(responseBody)), http.StatusInternalServerError)
		return
	}

	// 6. Парсим ответ и извлекаем текст
	var apiResponse OpenAIResponse
	if err := json.Unmarshal(responseBody, &apiResponse); err != nil {
		http.Error(w, fmt.Sprintf("Error unmarshalling API response: %v", err), http.StatusInternalServerError)
		return
	}

	var responseText string
	if len(apiResponse.Choices) > 0 {
		responseText = apiResponse.Choices[0].Message.Content
	} else {
		responseText = "Модель не вернула ответ."
	}
	log.Printf("Ответ модели: %s", responseText)

	// 7. Отправляем ответ обратно нашему фронтенду
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"response": responseText})
}

// serveIndex - обработчик для главной страницы
func serveIndex(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		http.Error(w, "Could not find index.html", http.StatusInternalServerError)
		log.Printf("Ошибка при парсинге index.html: %v", err)
		return
	}
	tmpl.Execute(w, nil)
}

func main() {
	// Регистрируем обработчики
	http.HandleFunc("/", serveIndex)
	http.HandleFunc("/api/chat", handleChat)

	port := "8000"
	log.Printf("Сервер запущен на http://localhost:%s", port)
	// Запускаем сервер
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Не удалось запустить сервер: %v", err)
	}
}