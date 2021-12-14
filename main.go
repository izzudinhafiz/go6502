package main

import (
	"fmt"

	cpu6502 "izzudinhafiz.com/go-6502/cpu"
)

func main(){
	fmt.Println("Hello World")
	c := cpu6502.New()
	c.ModifyMemory(1024*64 - 1, 1)
	c.PrintMemory()
}