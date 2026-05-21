// Тестовое приложение для проверки прохождения X-Forwarded-For через цепочку nginx.
// Отдаёт JSON с разобранным заголовком — цепочка хопов, длина, плюс сырые заголовки
// для отладки. Используется как конечный узел стенда docker-compose.
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
)

// response — структура ответа приложения.
// Поля сознательно дублируют сырые и разобранные значения, чтобы протокол тестирования
// мог проверять как строковое представление, так и количество элементов в цепочке.
type response struct {
	Path          string              `json:"path"`
	RemoteAddr    string              `json:"remote_addr"`     // TCP-пир: последний nginx перед приложением
	XForwardedFor string              `json:"x_forwarded_for"` // сырое значение заголовка как есть
	Chain         []string            `json:"chain"`           // та же цепочка, но разобранная по запятой
	ChainLength   int                 `json:"chain_length"`
	Headers       map[string][]string `json:"all_headers"` // все заголовки запроса — для отладки
}

func handler(w http.ResponseWriter, r *http.Request) {
	// X-Forwarded-For по RFC может приходить как одним заголовком с перечислением через запятую,
	// так и несколькими одноимёнными заголовками. r.Header.Values возвращает все значения,
	// объединяем их через ", " — получаем единую строку для дальнейшего парсинга.
	xff := strings.Join(r.Header.Values("X-Forwarded-For"), ", ")

	// Разбираем цепочку, отбрасывая пустые элементы (на случай двойных запятых
	// или висящих пробелов от какого-нибудь криво настроенного прокси).
	chain := make([]string, 0)
	for _, hop := range strings.Split(xff, ",") {
		hop = strings.TrimSpace(hop)
		if hop != "" {
			chain = append(chain, hop)
		}
	}

	resp := response{
		Path:          r.URL.Path,
		RemoteAddr:    r.RemoteAddr,
		XForwardedFor: xff,
		Chain:         chain,
		ChainLength:   len(chain),
		Headers:       r.Header,
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(resp); err != nil {
		// До сюда мы дойдём только если клиент закрыл соединение посреди записи —
		// логируем и идём дальше, что-то отвечать уже бессмысленно.
		log.Printf("ошибка записи ответа: %v", err)
	}
}

func main() {
	// Адрес прослушки задаётся через переменную окружения APP_ADDR — удобно для тестов
	// и для запуска вне Docker. По умолчанию слушаем все интерфейсы на 8080.
	addr := os.Getenv("APP_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)

	log.Printf("приложение готово, слушаем %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("сервер упал: %v", err)
	}
}
