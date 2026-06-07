package main

import (
	"strings"
	"unicode"

	"github.com/sirkon/errors"
)

type UserName string

func (u *UserName) UnmarshalText(text []byte) error {
	data := string(text)
	if !isValidUserToken(data) {
		return errors.New("invalid user name")
	}

	*u = UserName(data)
	return nil
}

// isValidUserToken проверяет токен по схеме: (A ( ('-' A)* ('-' Z)? )) | Z
func isValidUserToken(token string) bool {
	if token == "" {
		return false
	}

	// Особый случай: токен из одной части
	if !strings.Contains(token, "-") {
		// Как самостоятельный Z требует минимум 2 буквы
		return isZ(token, 2) || isA(token)
	}

	parts := strings.Split(token, "-")

	// Проверяем, что нет пустых частей
	for _, p := range parts {
		if p == "" {
			return false
		}
	}

	// Первая часть всегда должна быть A
	if !isA(parts[0]) {
		return false
	}

	// Проверяем средние части (все кроме первой и последней) — должны быть A
	for _, p := range parts[1 : len(parts)-1] {
		if !isA(p) {
			return false
		}
	}

	// Последняя часть может быть A или Z (Z в цепочке требует минимум 1 букву)
	lastPart := parts[len(parts)-1]
	return isA(lastPart) || isZ(lastPart, 1)
}

// isA проверяет, является ли строка токеном A: [a-z]{2,}
func isA(s string) bool {
	if len(s) < 2 {
		return false
	}
	for _, c := range s {
		if !unicode.IsLower(c) || !unicode.IsLetter(c) {
			return false
		}
	}
	return true
}

// isZ проверяет, является ли строка токеном Z
// minLetters - минимальное количество букв перед числом
func isZ(s string, minLetters int) bool {
	if len(s) < minLetters {
		return false
	}

	// Первый символ должен быть буквой
	if !unicode.IsLetter(rune(s[0])) || !unicode.IsLower(rune(s[0])) {
		return false
	}

	idx := 0
	for idx < len(s) && unicode.IsLetter(rune(s[idx])) {
		if !unicode.IsLower(rune(s[idx])) {
			return false
		}
		idx++
	}

	// Проверяем, что букв достаточно
	if idx < minLetters {
		return false
	}

	// Вся строка — буквы, это A.
	if idx == len(s) {
		return isA(s)
	}

	// Остаток должен быть положительным числом без ведущего нуля
	numPart := s[idx:]

	for _, c := range numPart {
		if !unicode.IsDigit(c) {
			return false
		}
	}

	// Первая цифра не ноль
	if numPart[0] == '0' {
		return false
	}

	return true
}
