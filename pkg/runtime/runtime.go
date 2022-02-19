package g

import "unsafe"

func getG() unsafe.Pointer

func GetG() unsafe.Pointer {
	return getG()
}
