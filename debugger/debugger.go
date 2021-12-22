package c6502debugger

import (
	"fmt"

	cpu "izzudinhafiz.com/go-6502/cpu"
)

type word uint16

type instructionPair struct {
	name string
	fetchSize int
}

type Trace struct {
	Op string
	Registers cpu.CpuRegisters
	Flags cpu.CpuFlags
	Clock int
	LastCycle int
	// MemAccess [][2]int
	Stack []byte
	NumOperations int
}

type Debugger6502 struct {
	cpu *cpu.Cpu6502
	TraceStack []Trace
	NumOperations int
	RecentWrite [][2]int
}

var INSTRUCTION_MAP = map[byte]instructionPair {
	cpu.ADR_IMMEDIATE: {"IMM", 1},
	cpu.ADR_ACCUMULATOR: {"ACM", 0},
	cpu.ADR_IMPLICIT: {"IMP", 0},
	cpu.ADR_ZEROPAGE: {"ZP0", 1},
	cpu.ADR_ZEROPAGEX: {"ZPX", 1},
	cpu.ADR_ZEROPAGEY: {"ZPY", 1},
	cpu.ADR_ABSOLUTE: {"ABS", 2},
	cpu.ADR_ABSOLUTEX: {"ABX", 2},
	cpu.ADR_ABSOLUTEY: {"ABY", 2},
	cpu.ADR_INDIRECT: {"IND", 2},
	cpu.ADR_INDIRECTX: {"INX", 1},
	cpu.ADR_INDIRECTY: {"INY", 1},
	cpu.ADR_RELATIVE: {"REL", 1},
}

func New(c *cpu.Cpu6502) *Debugger6502 {
	d := Debugger6502{cpu: c}
	return &d
}

func (d *Debugger6502) Trace() {
	op := d.DisassembleLine(int(d.cpu.Registers.PC))
	cycles := 0

	for !d.cpu.SingleStep() {
		cycles += 1
	}

	d.NumOperations += 1
	trace := Trace{op, d.cpu.Registers, d.cpu.Flags, d.cpu.Tick, cycles, d.getCPUStack(), d.NumOperations}
	d.TraceStack = append(d.TraceStack, trace)
}

func (d *Debugger6502) getCPUStack() []byte {
	return d.cpu.Memory[0x0101 + int(d.cpu.Registers.SP): 0x01FF + 1]
}

func (d *Debugger6502) DisassembleLine(startAddr int) string {
	var lo byte
	var hi byte
	var line string
	addr := startAddr

	ins_addr := addr
	op := d.cpu.Memory[addr]
	addr += 1

	opcode, key_exists := cpu.Opcodes[op]

	if !key_exists {
		return fmt.Sprintf("%#04X [XXX] INVALID OP", ins_addr)
	}

	instruction := INSTRUCTION_MAP[opcode.AddressingMode]
	addrMode, fetchSize := instruction.name, instruction.fetchSize

	line = fmt.Sprintf("%#04X [%v] %v ", ins_addr, addrMode, opcode.FriendlyName)

	if fetchSize == 0 {
	} else if fetchSize == 1 {
		lo = d.cpu.Memory[addr]
		addr += 1
		hi = 0

		if addrMode == "IMM" {
			line += "#"
			line += fmt.Sprintf("%#02X", lo)
		} else if addrMode == "REL" {
			line += fmt.Sprintf("%#02X", lo)

			rel := word(lo)
			if lo & 0x80 > 0 {
				rel |= 0xFF00
			}
			line += fmt.Sprintf("[$%#04X]", (word(addr) + rel))
		}
	} else {
		lo = d.cpu.Memory[addr]
		addr += 1
		hi = d.cpu.Memory[addr]
		fullAddr := word(hi) << 8 | word(lo)
		line += fmt.Sprintf("$%#04X", fullAddr)
	}

	if stringInSlice(addrMode, []string{"INX", "ZPX", "ABX"}) {
		line += ", X"
	} else if stringInSlice(addrMode, []string {"INY", "ZPY", "ABY"}){
		line += ", Y"
	}

	return line
}

func stringInSlice(s string, list []string) bool {
	for _, b := range list {
		if b == s {
			return true
		}
	}

	return false
}