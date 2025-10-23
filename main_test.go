package main

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCafeNegative(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		request string
		status  int
		message string
	}{
		{"/cafe", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=omsk", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=tula&count=na", http.StatusBadRequest, "incorrect count"},
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v.request, nil)
		handler.ServeHTTP(response, req)

		assert.Equal(t, v.status, response.Code)
		assert.Equal(t, v.message, strings.TrimSpace(response.Body.String()))
	}
}

func TestCafeWhenOk(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []string{
		"/cafe?count=2&city=moscow",
		"/cafe?city=tula",
		"/cafe?city=moscow&search=ложка",
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v, nil)

		handler.ServeHTTP(response, req)

		assert.Equal(t, http.StatusOK, response.Code)
	}
}

func TestCafeCount(t *testing.T) {

	requests := []struct {
		count int
		want  int
	}{
		{0, 0},
		{1, 1},
		{2, 2},
		{100, min(len(cafeList["moscow"]))},
	}
	for _, v := range requests {
		handler := http.HandlerFunc(mainHandle)
		response := httptest.NewRecorder()

		//Формирование запроса
		req := httptest.NewRequest("GET", "/cafe", nil)

		//Добавление к запросу параметров
		q := req.URL.Query()
		q.Add("city", "moscow")
		q.Add("count", strconv.Itoa(v.count))
		req.URL.RawQuery = q.Encode()

		//Запуск сервера со сформированным запросом
		handler.ServeHTTP(response, req)

		//Проверка успешности выполнения запроса
		require.Equal(t, http.StatusOK, response.Code)

		//Получение ответа
		assert.Equal(t, http.StatusOK, response.Code)
		res := response.Body.String()

		//Проверка на пустую строку и запись ответа в слайс
		var cafes []string
		if res != "" {
			cafes = strings.Split(res, ",")
		}

		//Сравнение результата
		assert.Len(t, len(cafes), v.want)
	}
}

func TestCafeSearch(t *testing.T) {
	requests := []struct {
		search    string
		wantCount int
	}{
		{"фасоль", 0},
		{"кофе", 2},
		{"вилка", 1},
	}
	for _, v := range requests {
		handler := http.HandlerFunc(mainHandle)
		response := httptest.NewRecorder()

		//Создание запроса и добавление параметров
		req := httptest.NewRequest("GET", "/cafe", nil)
		q := req.URL.Query()
		q.Add("city", "moscow")
		q.Add("search", v.search)
		req.URL.RawQuery = q.Encode()

		//Запуск сервера
		handler.ServeHTTP(response, req)

		//Получение ответа
		assert.Equal(t, http.StatusOK, response.Code)
		res := response.Body.String()

		//Проверка на пустую строку и запись ответа в слайс
		var cafes []string
		if res != "" {
			cafes = strings.Split(res, ",")
		}

		//Проверка количества найденных кафе
		assert.Len(t, len(cafes), v.wantCount)

		//Проверка содержания заданных символов
		searchLower := strings.ToLower(v.search)
		for _, cafe := range cafes {
			cafeLower := strings.ToLower(cafe)
			assert.Contains(t, cafeLower, searchLower)
		}
	}
}
