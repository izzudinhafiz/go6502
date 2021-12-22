package cpu6502

import (
	"fmt"
)

type word uint16
type CpuFlags struct {
	N byte
	Z byte
	C byte
	I byte
	D byte
	B byte
	V byte
}

type CpuRegisters struct {
	PC word
	SP byte
	A byte
	X byte
	Y byte
}

type Cpu6502 struct {
	Clock int
	Fetched byte
	AbsoluteAddr word
	RelativeAddr word
	Registers CpuRegisters
	Flags CpuFlags
	Opcode Opcode
	Memory []byte
	Tick int
}

func New() *Cpu6502 {
	c := Cpu6502{Memory: make([]byte, 1024*64)}
	c.Reset()

	return &c
}

func (c *Cpu6502) Reset() {
	c.Fetched = 0
	c.AbsoluteAddr = 0
	c.RelativeAddr = 0
	c.Opcode = Opcode{1, "NOP", OP_NOP, ADR_IMPLICIT, nop, implicit}
	c.Registers.SP = 0xFD
	c.Registers.A = 0
	c.Registers.X = 0
	c.Registers.Y = 0

	c.Flags.N = 0
	c.Flags.Z = 0
	c.Flags.C = 0
	c.Flags.I = 0
	c.Flags.D = 0
	c.Flags.B = 0
	c.Flags.V = 0

	c.Registers.PC = word(c.Memory[0xFFFD]) << 8 | word(c.Memory[0xFFFC])
	c.Clock = 0
	c.Tick = 0
}

func (c *Cpu6502) SetResetVector(addr word){
	c.Memory[0xFFFD] = byte(addr >> 8)
	c.Memory[0xFFFC] = byte(addr)
}

// Runs one clock cycle of the CPU, returns true when an operation has just been completed
func (c *Cpu6502) SingleStep() bool {
	c.Tick += 1
	if c.Clock == 0 {
		current_byte := c.fetchByte()
		current_op, key_exists := Opcodes[current_byte]
		if !key_exists {
			panic("UNKNOWN OPCODE READ")
		}

		c.Opcode = current_op
		c.Clock += c.Opcode.NumCycle

		c.Clock += c.Opcode.Address(c)
		c.Clock += c.Opcode.Op(c)
	} else {
		c.Clock -= 1
	}

	if c.Clock > 0 { return true }
	return false
}

// Runs a single opcode to completion
func (c *Cpu6502) SingleOperation() {
	for {
		if c.SingleStep() { break }
	}
}

func (c *Cpu6502) GetStatusFlags() CpuFlags {
	return CpuFlags{c.Flags.N, c.Flags.Z, c.Flags.C, c.Flags.I, c.Flags.D, c.Flags.B, c.Flags.V}
}

func (c *Cpu6502) getStatusFlagsByte(mode string) byte {
	var temp byte
	if mode == "instruction" {
		temp = (c.Flags.C << 0) | (c.Flags.Z << 1) | (c.Flags.I << 2) | (c.Flags.D << 3) | (byte(1) << 4) | (byte(1) << 5) | (c.Flags.V << 6) | (c.Flags.N << 7)
	} else {
		temp = (c.Flags.C << 0) | (c.Flags.Z << 1) | (c.Flags.I << 2) | (c.Flags.D << 3) | (byte(0) << 4) | (byte(1) << 5) | (c.Flags.V << 6) | (c.Flags.N << 7)
	}

	return temp
}

func (c *Cpu6502) setStatusFlags(value byte) {
	if value & (1 << 0) > 0 {
		c.Flags.C = 1
	} else {
		c.Flags.C = 0
	}

	if value & (1 << 1) > 0 {
		c.Flags.Z = 1
	} else {
		c.Flags.Z = 0
	}

	if value & (1 << 2) > 0 {
		c.Flags.I = 1
	} else {
		c.Flags.I = 0
	}

	if value & (1 << 3) > 0 {
		c.Flags.D = 1
	} else {
		c.Flags.D = 0
	}

	if value & (1 << 4) > 0 {
		c.Flags.B = 1
	} else {
		c.Flags.B = 0
	}

	if value & (1 << 6) > 0 {
		c.Flags.V = 1
	} else {
		c.Flags.V = 0
	}

	if value & (1 << 7) > 0 {
		c.Flags.N = 1
	} else {
		c.Flags.N = 0
	}

}

func (c *Cpu6502) setNZFlag(val byte){
	// If value is zero
	if val == 0 {
		c.Flags.Z = 1
	} else {
		c.Flags.Z = 0
	}

	// if value is negative
	if val & (1 << 7) > 0 {
		c.Flags.N = 1
	} else {
		c.Flags.N = 0
	}
}

func (c *Cpu6502) fetchByte() byte {
	value := c.Memory[c.Registers.PC]
	c.Registers.PC += 1
	return value
}

func (c *Cpu6502) fetchWord() word {
	lo_byte := word(c.fetchByte())
	hi_byte := word(c.fetchByte()) << 8
	return hi_byte | lo_byte
}

func (c *Cpu6502) WriteWord(data word, addr word){
	c.Memory[addr] = byte(data | 0xFF)
	c.Memory[addr + 1] = byte(data >> 8)
}

func (c *Cpu6502) ReadWord(addr word) word {
	lo_byte := word(c.Memory[addr])
	hi_byte := word(c.Memory[addr + 1]) << 8
	return hi_byte | lo_byte
}

func (c *Cpu6502) WriteMemory(startAddr int, data []byte){
	for i, _byte := range data {
		c.Memory[startAddr + i] = _byte
	}
}

func (c *Cpu6502) fetch() byte {
	if c.Opcode.AddressingMode != ADR_ACCUMULATOR{
		c.Fetched = c.Memory[c.AbsoluteAddr]
	}

	return c.Fetched
}

func (c *Cpu6502) stackPush(value byte){
	c.Memory[0x0100 + uint16(c.Registers.SP)] = value
	c.Registers.SP = (c.Registers.SP - 1) & 0xFF
}

func (c *Cpu6502) stackPull() byte {
	c.Registers.SP = (c.Registers.SP + 1) & 0xFF
	return c.Memory[0x0100 + uint16(c.Registers.SP)]
}

func (c *Cpu6502) Print() {
	fmt.Printf("Clock = %v, fetched = %v, absoluteAddr = %v, ", c.Clock, c.Fetched, c.AbsoluteAddr)
	fmt.Printf("PC = %#04x\n", c.Registers.PC)
}