package validator

import (
	"regexp"
	"strings"
	"unicode"
)

var validate *Validator

func init() {
	validate = &Validator{}
}

type Validator struct{}

func ValidateStruct(s interface{}) error {
	// Простая валидация для демонстрации
	// В реальном проекте используйте github.com/go-playground/validator/v10
	return nil
}

func validatePassword(password string) bool {
	// Минимальная длина 8 символов
	if len(password) < 8 {
		return false
	}

	// Проверяем наличие различных типов символов
	var (
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	// Требуем минимум 3 из 4 типов символов
	count := 0
	if hasUpper {
		count++
	}
	if hasLower {
		count++
	}
	if hasNumber {
		count++
	}
	if hasSpecial {
		count++
	}

	return count >= 3
}

func validateEmail(email string) bool {
	// Базовая проверка формата email
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return false
	}

	// Проверяем длину
	if len(email) > 254 {
		return false
	}

	// Проверяем, что email не содержит пробелов
	return !strings.Contains(email, " ")
}
