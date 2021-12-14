package cpu6502

import (
	"fmt"
	"reflect"
)

type cpuFlags struct {
	N uint8
	Z uint8
	C uint8
	I uint8
	D uint8
	B uint8
	V uint8
}

type cpuRegisters struct {
	PC uint8
	SP uint8
	A uint8
	X uint8
	Y uint8
}

type operation func(*cpu6502) int
type addresingMode func(*cpu6502) int

type operations struct {
	numCycle int
	op operation
	addrMode addresingMode
}

type cpu6502 struct {
	clock int
	fetched uint8
	absoluteAddr uint8
	relativeAddr uint8
	registers cpuRegisters
	flags cpuFlags
	opcode operations
	memory []uint8
	lookupTable map[uint8]operations
}

func New() *cpu6502 {
	c := cpu6502{
		registers: cpuRegisters{},
		flags: cpuFlags{},
		memory: make([]uint8, 1024*64),
		opcode: operations{1, nop, implicit},
		lookupTable: map[uint8]operations{
			0x69: {2, adc, accumulator},
			0x65: {3, adc, zeropage},
		},
	}
	return &c
}

func (c cpu6502) PrintMemory() {
	fmt.Println(c.memory)
}

func (c cpu6502) ModifyMemory(idx int, value uint8) {
	c.memory[idx] = value
}

func (c cpu6502) FetchByte() uint8 {
	value := c.memory[c.registers.PC]
	c.registers.PC += 1
	return value
}

func (c cpu6502) FetchWord() uint16 {
	lo_byte := uint16(c.FetchByte())
	hi_byte := uint16(c.FetchByte()) << 8
	return hi_byte | lo_byte
}

func (c cpu6502) WriteWord(data uint16, addr int){
	c.memory[addr] = uint8(data | 0xFF)
	c.memory[addr + 1] = uint8(data >> 8)
}

func (c cpu6502) ReadWord(addr int) uint16 {
	lo_byte := uint16(c.memory[addr])
	hi_byte := uint16(c.memory[addr + 1]) << 8
	return hi_byte | lo_byte
}

//TODO: This dont work. Need a different way to compare two functions
func (c cpu6502) Fetch() uint8 {
	if reflect.ValueOf(c.opcode.addrMode) != reflect.ValueOf(accumulator){
		c.fetched = c.memory[c.absoluteAddr]
	}

	return c.fetched
}

func accumulator(c *cpu6502) int {
	c.fetched = c.registers.A
	return 2
}

func implicit(c *cpu6502) int {
	return 1
}

func zeropage(c *cpu6502) int {
	c.absoluteAddr = c.FetchByte() + c.registers.X
	return 1
}

func adc(c *cpu6502) int {
	return 1
}

func nop(c *cpu6502) int {
	return 1
}

