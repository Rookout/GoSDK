package utils

import (
	"container/list"
	"strings"
)

func GetElementInList(l *list.List, index int) interface{} {
	i := 0
	for e := l.Front(); e != nil; e = e.Next() {
		if index == i {
			return e.Value
		}
		i++
	}
	return nil
}

func Contains(slice []string, str string) bool {
	for _, value := range slice {
		if value == str {
			return true
		}
	}
	return false
}

func IsTrue(str string) bool {
	return Contains(TrueValues, strings.ToLower(str))
}

func Cut(str string, l int) string {
	if len(str) > l {
		return str[:l]
	}
	return str
}
