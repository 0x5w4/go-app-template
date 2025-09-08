package util

import (
	"html"
	"reflect"
	"strings"
)

func Sanitize(data any) {
	val := reflect.ValueOf(data).Elem()

	for i := range val.NumField() {
		field := val.Field(i)

		if !field.CanSet() {
			continue
		}

		switch field.Kind() {
		case reflect.String:
			originalString := field.String()

			// 1. Lakukan TrimSpace
			trimmedString := strings.TrimSpace(originalString)

			// 2. Lakukan html.EscapeString (mirip htmlspecialchars)
			sanitizedString := html.EscapeString(trimmedString)

			// Set nilai field dengan string yang sudah bersih
			field.SetString(sanitizedString)

		case reflect.Struct:
			// Jika field adalah struct, panggil Sanitize secara rekursif
			// Kita perlu mengambil alamat dari struct ini untuk bisa memodifikasinya
			if field.Addr().CanInterface() {
				Sanitize(field.Addr().Interface())
			}

		case reflect.Ptr:
			// Jika field adalah pointer ke sesuatu
			if !field.IsNil() && field.Elem().Kind() == reflect.Struct {
				// Dan itu adalah pointer ke struct, panggil Sanitize secara rekursif
				Sanitize(field.Interface())
			}
		}
	}
}
