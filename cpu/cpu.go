package cpu6502

import (
	"fmt"
)

type word uint16
type cpuFlags struct {
	N byte
	Z byte
	C byte
	I byte
	D byte
	B byte
	V byte
}

type cpuRegisters struct {
	PC word
	SP byte
	A byte
	X byte
	Y byte
}

type cpu6502 struct {
	clock int
	fetched byte
	absoluteAddr word
	relativeAddr word
	registers cpuRegisters
	flags cpuFlags
	opcode Opcode
	Memory []byte
	lookupTable map[byte]Opcode
}

func New() *cpu6502 {
	c := cpu6502{Memory: make([]byte, 1024*64), lookupTable: Opcodes}
	c.Reset()

	return &c
}

func (c *cpu6502) Reset() {
	c.fetched = 0
	c.absoluteAddr = 0
	c.relativeAddr = 0
	c.opcode = Opcode{1, OP_NOP, ADR_IMPLICIT, nop, implicit}
	c.registers.SP = 0xFD
	c.registers.A = 0
	c.registers.X = 0
	c.registers.Y = 0

	c.flags.N = 0
	c.flags.Z = 0
	c.flags.C = 0
	c.flags.I = 0
	c.flags.D = 0
	c.flags.B = 0
	c.flags.V = 0

	c.registers.PC = word(c.Memory[0xFFFD]) << 8 | word(c.Memory[0xFFFC])
	c.clock = 0
}

func (c *cpu6502) SetResetVector(addr word){
	c.Memory[0xFFFD] = byte(addr >> 8)
	c.Memory[0xFFFC] = byte(addr)
}

func (c *cpu6502) SingleStep() bool {
	if c.clock == 0 {
		current_byte := c.fetchByte()
		current_op, key_exists := c.lookupTable[current_byte]
		if !key_exists {
			panic("UNKNOWN OPCODE READ")
		}

		c.opcode = current_op
		c.clock += c.opcode.NumCycle

		c.clock += c.opcode.Address(c)
		c.clock += c.opcode.Op(c)
	} else {
		c.clock -= 1
	}

	if c.clock > 0 { return true }
	return false
}

func (c *cpu6502) SingleOperation() {
	for {
		if c.SingleStep() { break }
	}
}

func (c *cpu6502) GetStatusFlags() cpuFlags {
	return cpuFlags{c.flags.N, c.flags.Z, c.flags.C, c.flags.I, c.flags.D, c.flags.B, c.flags.V}
}

func (c *cpu6502) getStatusFlagsByte(mode string) byte {
	var temp byte
	if mode == "instruction" {
		temp = (c.flags.C << 0) | (c.flags.Z << 1) | (c.flags.I << 2) | (c.flags.D << 3) | (byte(1) << 4) | (byte(1) << 5) | (c.flags.V << 6) | (c.flags.N << 7)
	} else {
		temp = (c.flags.C << 0) | (c.flags.Z << 1) | (c.flags.I << 2) | (c.flags.D << 3) | (byte(0) << 4) | (byte(1) << 5) | (c.flags.V << 6) | (c.flags.N << 7)
	}

	return temp
}

func (c *cpu6502) setStatusFlags(value byte) {
	if value & (1 << 0) > 0 {
		c.flags.C = 1
	} else {
		c.flags.C = 0
	}

	if value & (1 << 1) > 0 {
		c.flags.Z = 1
	} else {
		c.flags.Z = 0
	}

	if value & (1 << 2) > 0 {
		c.flags.I = 1
	} else {
		c.flags.I = 0
	}

	if value & (1 << 3) > 0 {
		c.flags.D = 1
	} else {
		c.flags.D = 0
	}

	if value & (1 << 4) > 0 {
		c.flags.B = 1
	} else {
		c.flags.B = 0
	}

	if value & (1 << 6) > 0 {
		c.flags.V = 1
	} else {
		c.flags.V = 0
	}

	if value & (1 << 7) > 0 {
		c.flags.N = 1
	} else {
		c.flags.N = 0
	}

}

func (c *cpu6502) setNZFlag(val byte){
	// If value is zero
	if val == 0 {
		c.flags.Z = 1
	} else {
		c.flags.Z = 0
	}

	// if value is negative
	if val & (1 << 7) > 0 {
		c.flags.N = 1
	} else {
		c.flags.N = 0
	}
}

func (c *cpu6502) fetchByte() byte {
	value := c.Memory[c.registers.PC]
	c.registers.PC += 1
	return value
}

func (c *cpu6502) fetchWord() word {
	lo_byte := word(c.fetchByte())
	hi_byte := word(c.fetchByte()) << 8
	return hi_byte | lo_byte
}

func (c *cpu6502) WriteWord(data word, addr word){
	c.Memory[addr] = byte(data | 0xFF)
	c.Memory[addr + 1] = byte(data >> 8)
}

func (c *cpu6502) ReadWord(addr word) word {
	lo_byte := word(c.Memory[addr])
	hi_byte := word(c.Memory[addr + 1]) << 8
	return hi_byte | lo_byte
}

func (c *cpu6502) WriteMemory(startAddr int, data []byte){
	for i, _byte := range data {
		c.Memory[startAddr + i] = _byte
	}
}

func (c *cpu6502) fetch() byte {
	if c.opcode.AddressingMode != ADR_ACCUMULATOR{
		c.fetched = c.Memory[c.absoluteAddr]
	}

	return c.fetched
}

func (c *cpu6502) stackPush(value byte){
	c.Memory[0x0100 + uint16(c.registers.SP)] = value
	c.registers.SP = (c.registers.SP - 1) & 0xFF
}

func (c *cpu6502) stackPull() byte {
	c.registers.SP = (c.registers.SP + 1) & 0xFF
	return c.Memory[0x0100 + uint16(c.registers.SP)]
}

func (c *cpu6502) Print() {
	fmt.Printf("Clock = %v, fetched = %v, absoluteAddr = %v, ", c.clock, c.fetched, c.absoluteAddr)
	fmt.Printf("PC = %#04x\n", c.registers.PC)
}