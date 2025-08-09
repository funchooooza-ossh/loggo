package formatter

import (
	"reflect"
	"strconv"
	"strings"
)

func toFloatString(v interface{}) string {
	switch f := v.(type) {
	case float32:
		return strconv.FormatFloat(float64(f), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(f, 'f', -1, 64)
	default:
		return `"invalid_float"`
	}
}

// Возвращает ok=false, если rv уже встречался в текущем стеке обхода.
// release() нужно вызвать при выходе из узла (обычно через defer).
func markAndCheck(rv reflect.Value, visited map[uintptr]struct{}) (ok bool, release func()) {
	switch rv.Kind() {
	case reflect.Ptr, reflect.Slice, reflect.Map, reflect.Func, reflect.Chan, reflect.UnsafePointer:
		p := rv.Pointer()
		if p != 0 {
			if _, seen := visited[p]; seen {
				return false, func() {}
			}
			visited[p] = struct{}{}
			return true, func() { delete(visited, p) }
		}
	case reflect.Struct:
		if rv.CanAddr() {
			p := rv.Addr().Pointer()
			if p != 0 {
				if _, seen := visited[p]; seen {
					return false, func() {}
				}
				visited[p] = struct{}{}
				return true, func() { delete(visited, p) }
			}
		}
	}
	return true, func() {}
}

// addMultilinePrefix вставляет префикс "│ " после каждого перевода строки.
// Пример: "a\nb" -> "a\n│ b"
func addMultilinePrefix(s string) string {
	// нормализуем CRLF -> LF, затем вставляем префикс
	if strings.IndexByte(s, '\n') == -1 && !strings.Contains(s, "\r\n") {
		return s
	}
	s = strings.ReplaceAll(s, "\r\n", "\n")
	return strings.ReplaceAll(s, "\n", "\n│ ")
}
