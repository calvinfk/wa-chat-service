package main

import (
	"encoding/base64"
)

func main() {
	data := ""
	encoded := base64.StdEncoding.EncodeToString([]byte(data))
	println(encoded)
}
