package main

import "fmt"

func main() {
	t := "asdasdassda\\n"

	fmt.Println(t[:len(t)-1])
}
