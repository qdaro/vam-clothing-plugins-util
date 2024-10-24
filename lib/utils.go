package lib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
)

func Check(err error) {
	if err != nil {
		panic(err)
	}
}

func Must[T any](t T, err error) T {
	Check(err)
	return t
}

func Ptr[T any](value T) *T {
	return &value
}

func Find[T any](slice []T, predicate func(T) bool) (T, bool) {
	for _, v := range slice {
		if predicate(v) {
			return v, true
		}
	}
	var zero T
	return zero, false
}

func ReadJSON[T any](path string) (T, error) {
	var result T

	data, err := os.ReadFile(path)
	if err != nil {
		return result, err
	}

	err = json.Unmarshal(data, &result)
	if err != nil {
		return result, err
	}

	return result, nil
}

func IsType[T any](value interface{}) bool {
	_, ok := value.(T)
	return ok
}

func JSONMarshalPretty(t any) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "\t")
	err := encoder.Encode(t)
	return buffer.Bytes(), err
}

func JSONMarshalLog(t any) string {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	err := encoder.Encode(t)
	if err != nil {
		return fmt.Sprintf("Couldn't encode json data. Error: %v", err)
	}
	return buffer.String()
}

// Ensures the structure is []map[string]string
func ParseToStringStringMap(data interface{}) (map[string]string, bool) {
	rawMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, false
	}

	strMap := map[string]string{}
	for key, value := range rawMap {
		strVal, ok := value.(string)
		if !ok {
			return nil, false
		}
		strMap[key] = strVal
	}

	return strMap, true
}

func FindIndexFromOffset(re *regexp.Regexp, s []byte, offset int) []int {
	sub := s[offset:]

	match := re.FindIndex(sub)
	if match == nil {
		return nil
	}

	return []int{match[0] + offset, match[1] + offset}
}
