package helpers

import (
	"reflect"
	"strings"
	"fmt"
)

func UniqueInt(m map[int]int) map[int]int {
	n := make(map[int]int, len(m))
	ref := make(map[int]bool, len(m))
	for k, v := range m {
		if _, ok := ref[v]; !ok {
			ref[v] = true
			n[k] = v
		}
	}
	return n
}

func InArray(val interface{}, array interface{}) (exists bool, index int) {
	exists = false
	index = -1

	switch reflect.TypeOf(array).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(array)

		for i := 0; i < s.Len(); i++ {
			if reflect.DeepEqual(val, s.Index(i).Interface()) == true {
				index = i
				exists = true
				return
			}
		}
	}

	return
}

func IsEmptyString(val string) bool {
	fmt.Println("--------------------")
	fmt.Println(val)
	return strings.Trim(val, " ") == ""
}

func IsEmptyNumber(val int64) bool {
	return val == 0
}
