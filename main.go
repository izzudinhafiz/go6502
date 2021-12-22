package main

import (
	"fmt"
	"os"

	cpu6502 "izzudinhafiz.com/go-6502/cpu"
	debugger "izzudinhafiz.com/go-6502/debugger"
)

func main(){
	f, err := os.Open("6502_functional_test.bin")
	if err != nil {
		panic(err)
	}
	stats, statsErr := f.Stat()
	if statsErr != nil {
		panic(statsErr)
	}
	size := stats.Size()
	readBuffer := make([]byte, size)
	f.Read(readBuffer)
	f.Close()

	fmt.Println(len(readBuffer))
	cpu := cpu6502.New()
	deb := debugger.New(cpu)

	cpu.WriteMemory(0, readBuffer)
	fmt.Println(cpu.Memory[0x400 - 10: 0x400+10])
	cpu.SetResetVector(0x400)
	cpu.Reset()

	repeat := 1000
	for i := 0; i < repeat; i++ {
		deb.Trace()
	}

	for i := 0; i < repeat; i++ {
		fmt.Println(deb.TraceStack[i])
	}
}