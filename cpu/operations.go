package cpu6502

func twos_comp(val int) int {
	if val&(1<<7) > 0 {
		val = val - 256
	}

	return val
}

func IRQ(c *cpu6502) int {
	if c.flags.I == 0 {
		c.stackPush(byte(c.registers.PC >> 8))
		c.stackPush(byte(c.registers.PC))

		c.flags.I = 1
		c.stackPush(c.getStatusFlagsByte("interrupt"))

		c.registers.PC = (word(c.Memory[0xFFFF]) << 8) | word(c.Memory[0xFFFE])

		return 7
	}
	return 0
}

func NMI(c *cpu6502) int {
	c.stackPush(byte(c.registers.PC >> 8))
	c.stackPush(byte(c.registers.PC))

	c.flags.I = 1
	c.stackPush(c.getStatusFlagsByte("interrupt"))
	c.registers.PC = (word(c.Memory[0xFFFB]) << 8) | word(c.Memory[0xFFFA])
	return 8
}

func nop(c *cpu6502) int {
	return 0
}

func adc(c *cpu6502) int {
	// Add with carry operation
	// Decimal mode: N, V, Z flags are invalid
	val := c.fetchByte()

	// Binary mode
	if c.flags.D == 0 {
		total := word(c.registers.A) + word(val) + word(c.flags.C)

		// Overflow check
		// See http://www.righto.com/2012/12/the-6502-overflow-flag-explained.html
		if ((word(val) ^ total) & (word(c.registers.A) ^ total) & 0x80) != 0 {
			c.flags.V = 1
		} else {
			c.flags.V = 0
		}

		if total > 0xFF {
			c.flags.C = 1
		} else {
			c.flags.C = 0
		}

		c.registers.A = byte(total)
		c.setNZFlag(c.registers.A)

		return 0
	} else {
		// Decimal mode
		a := int(c.registers.A)
		b := int(val)
		temp := (a & 0x0F) + (b & 0x0F) + int(c.flags.C) // Add 1s place of decimal

		if temp >= 0x0A { // If bigger than 9
			temp = ((temp + 0x06) & 0x0F) + 0x10 // Add 6 to skip 10 -> 15, keep the 1s place and carry to 10s place
		}

		a = (a & 0xF0) + (b & 0xF0) + temp // Add the 10s place in decimal

		if (a&0xFF)&(1<<7) > 0 { // N set if 7th bit is set
			c.flags.N = 1
		} else {
			c.flags.N = 0
		}

		if -128 <= twos_comp(a&0xFF) && twos_comp(a&0xFF) <= 127 {
			c.flags.V = 0
		} else {
			c.flags.V = 1
		}

		if a >= 0xA0 { // If bigger than 100
			a = a + 0x60 // skip 1xx -> 5xx, keeps the 10s and 1s place
		}

		c.registers.A = byte(a)
		if a >= 0x100 {
			c.flags.C = 1
		} else {
			c.flags.C = 0
		}

		return 0
	}
}

func and(c *cpu6502) int {
	val := c.fetch()
	c.registers.A = val & c.registers.A
	c.setNZFlag(c.registers.A)

	return 0
}

func asl(c *cpu6502) int {
	var val word
	if c.opcode.AddressingMode == ADR_ACCUMULATOR {
		val = word(c.registers.A) << 1
		c.registers.A = byte(val)
	} else {
		val = word(c.fetch()) << 1
		c.Memory[c.absoluteAddr] = byte(val)
	}

	c.flags.C = 0
	if val > 255 {
		c.flags.C = 1
	}
	c.setNZFlag(byte(val))
	return 0
}

func bcc(c *cpu6502) int {
	if c.flags.C == 0 {
		cycles := 1
		c.absoluteAddr = (c.registers.PC + c.relativeAddr)

		// If branch to new page, add a cycle
		if c.absoluteAddr&0xFF00 != c.registers.PC&0xFF00 {
			cycles += 1
		}
		c.registers.PC = c.absoluteAddr

		return cycles
	}
	return 0
}

func bcs(c *cpu6502) int {
	if c.flags.C == 1 {
		cycles := 1
		c.absoluteAddr = (c.registers.PC + c.relativeAddr)

		// If branch to new page, add a cycle
		if c.absoluteAddr&0xFF00 != c.registers.PC&0xFF00 {
			cycles += 1
		}
		c.registers.PC = c.absoluteAddr

		return cycles
	}
	return 0
}

func beq(c *cpu6502) int {
	if c.flags.Z == 1 {
		cycles := 1
		c.absoluteAddr = (c.registers.PC + c.relativeAddr)

		// If branch to new page, add a cycle
		if c.absoluteAddr&0xFF00 != c.registers.PC&0xFF00 {
			cycles += 1
		}
		c.registers.PC = c.absoluteAddr

		return cycles
	}
	return 0
}

func bit(c *cpu6502) int {
	val := c.fetch()
	c.flags.Z = 0
	c.flags.N = 0
	c.flags.V = 0

	if c.registers.A&val == 0 {
		c.flags.Z = 1
	}
	if val&(1<<7) > 0 {
		c.flags.N = 1
	}
	if val&(1<<6) > 0 {
		c.flags.V = 1
	}

	return 0
}

func bmi(c *cpu6502) int {
	if c.flags.N == 1 {
		cycles := 1
		c.absoluteAddr = (c.registers.PC + c.relativeAddr)

		// If branch to new page, add a cycle
		if c.absoluteAddr&0xFF00 != c.registers.PC&0xFF00 {
			cycles += 1
		}
		c.registers.PC = c.absoluteAddr

		return cycles
	}
	return 0
}

func bne(c *cpu6502) int {
	if c.flags.Z == 0 {
		cycles := 1
		c.absoluteAddr = (c.registers.PC + c.relativeAddr)

		// If branch to new page, add a cycle
		if c.absoluteAddr&0xFF00 != c.registers.PC&0xFF00 {
			cycles += 1
		}
		c.registers.PC = c.absoluteAddr

		return cycles
	}
	return 0
}

func bpl(c *cpu6502) int {
	if c.flags.N == 0 {
		cycles := 1
		c.absoluteAddr = (c.registers.PC + c.relativeAddr)

		// If branch to new page, add a cycle
		if c.absoluteAddr&0xFF00 != c.registers.PC&0xFF00 {
			cycles += 1
		}
		c.registers.PC = c.absoluteAddr

		return cycles
	}
	return 0
}

func brk(c *cpu6502) int {
	c.registers.PC += 1
	c.stackPush(byte(c.registers.PC >> 8)) // write high byte
	c.stackPush(byte(c.registers.PC))      // write low byte

	// On a BRK instruction, the CPU does the same as in the IRQ case,
	// but sets bit #4 (B flag) IN THE COPY of the status register that is saved on the stack.
	c.stackPush(c.getStatusFlagsByte("instruction"))
	c.flags.I = 1

	c.registers.PC = word(c.Memory[0xFFFF])<<8 | word(c.Memory[0xFFFE])

	return 0
}

func bvc(c *cpu6502) int {
	if c.flags.V == 0 {
		cycles := 1
		c.absoluteAddr = (c.registers.PC + c.relativeAddr)

		// If branch to new page, add a cycle
		if c.absoluteAddr&0xFF00 != c.registers.PC&0xFF00 {
			cycles += 1
		}
		c.registers.PC = c.absoluteAddr

		return cycles
	}
	return 0
}

func bvs(c *cpu6502) int {
	if c.flags.V == 1 {
		cycles := 1
		c.absoluteAddr = (c.registers.PC + c.relativeAddr)

		// If branch to new page, add a cycle
		if c.absoluteAddr&0xFF00 != c.registers.PC&0xFF00 {
			cycles += 1
		}
		c.registers.PC = c.absoluteAddr

		return cycles
	}
	return 0
}

func clc(c *cpu6502) int {
	c.flags.C = 0
	return 0
}

func cld(c *cpu6502) int {
	c.flags.D = 0
	return 0
}

func cli(c *cpu6502) int {
	c.flags.I = 0
	return 0
}

func clv(c *cpu6502) int {
	c.flags.V = 0
	return 0
}

func cmp(c *cpu6502) int {
	val := c.fetch()

	c.flags.C = 0
	if c.registers.A >= val {
		c.flags.C = 1
	}
	c.setNZFlag(c.registers.A - val)
	return 0
}

func cpx(c *cpu6502) int {
	val := c.fetch()

	c.flags.C = 0
	if c.registers.X >= val {
		c.flags.C = 1
	}
	c.setNZFlag(c.registers.X - val)
	return 0
}

func cpy(c *cpu6502) int {
	val := c.fetch()

	c.flags.C = 0
	if c.registers.Y >= val {
		c.flags.C = 1
	}
	c.setNZFlag(c.registers.Y - val)
	return 0
}

func dec(c *cpu6502) int {
	val := c.fetch() - 1
	c.Memory[c.absoluteAddr] = val
	c.setNZFlag(val)
	return 0
}

func dex(c *cpu6502) int {
	c.registers.X -= 1
	c.setNZFlag(c.registers.X)
	return 0
}

func dey(c *cpu6502) int {
	c.registers.Y -= 1
	c.setNZFlag(c.registers.Y)
	return 0
}

func eor(c *cpu6502) int {
	c.registers.A = c.registers.A ^ c.fetch()
	c.setNZFlag(c.registers.A)
	return 0
}

func inc(c *cpu6502) int {
	val := c.fetch() + 1
	c.Memory[c.absoluteAddr] = val
	c.setNZFlag(val)
	return 0
}

func inx(c *cpu6502) int {
	c.registers.X = c.registers.X + 1
	c.setNZFlag(c.registers.X)
	return 0
}

func iny(c *cpu6502) int {
	c.registers.Y = c.registers.Y + 1
	c.setNZFlag(c.registers.Y)
	return 0
}

func jmp(c *cpu6502) int {
	c.registers.PC = c.absoluteAddr
	return 0
}

func jsr(c *cpu6502) int {
	c.registers.PC -= 1
	c.stackPush(byte(c.registers.PC >> 8)) // Write high byte
	c.stackPush(byte(c.registers.PC))      // write low byte

	c.registers.PC = c.absoluteAddr
	return 0
}

func lda(c *cpu6502) int {
	c.registers.A = c.fetch()
	c.setNZFlag(c.registers.A)
	return 0
}

func ldx(c *cpu6502) int {
	c.registers.X = c.fetch()
	c.setNZFlag(c.registers.X)
	return 0
}

func ldy(c *cpu6502) int {
	c.registers.Y = c.fetch()
	c.setNZFlag(c.registers.Y)
	return 0
}

func lsr(c *cpu6502) int {
	var val byte
	if c.opcode.AddressingMode == ADR_ACCUMULATOR {
		val = c.registers.A
	} else {
		val = c.fetch()
	}
	c.Memory[c.absoluteAddr] = val >> 1

	c.flags.C = 0
	if val&0x01 > 0 {
		c.flags.C = 1
	}
	c.setNZFlag(c.flags.C)
	return 0
}

func ora(c *cpu6502) int {
	c.registers.A = c.registers.A | c.fetch()
	c.setNZFlag(c.registers.A)
	return 0
}

func pha(c *cpu6502) int {
	c.stackPush(c.registers.A)
	return 0
}

func pla(c *cpu6502) int {
	c.registers.A = c.stackPull()
	c.setNZFlag(c.registers.A)
	return 0
}

func php(c *cpu6502) int {
	c.stackPush(c.getStatusFlagsByte("instruction"))
	c.flags.B = 0
	return 0
}

func plp(c *cpu6502) int {
	c.setStatusFlags(c.stackPull())
	c.flags.B = 0
	return 0
}

func rol(c *cpu6502) int {
	var val word
	var temp byte
	if c.opcode.AddressingMode == ADR_ACCUMULATOR {
		val = word(c.registers.A)
		temp := byte(val << 1) | c.flags.C
		c.registers.A = temp
	} else {
		val = word(c.fetch())
		temp := byte(val << 1) | c.flags.C
		c.Memory[c.absoluteAddr] = temp
	}

	c.flags.C = 0
	if val & 0x01 > 0 { c.flags.C = 1}
	c.setNZFlag(temp)
	return 0
}

func ror(c *cpu6502) int {
	var val word
	var temp byte
	if c.opcode.AddressingMode == ADR_ACCUMULATOR {
		val = word(c.registers.A)
		temp := byte(val >> 1) | (c.flags.C << 7)
		c.registers.A = temp
	} else {
		val = word(c.fetch())
		temp := byte(val >> 1) | (c.flags.C << 7)
		c.Memory[c.absoluteAddr] = temp
	}

	c.flags.C = 0
	if val & 0x01 > 0 { c.flags.C = 1}
	c.setNZFlag(temp)
	return 0
}

func rti(c *cpu6502) int {
	c.setStatusFlags(c.stackPull())
	c.flags.B = 0

	lo_byte := c.stackPull()
	hi_byte := c.stackPull()

	c.registers.PC = word(hi_byte) << 8 | word(lo_byte)
	return 0
}

func rts(c *cpu6502) int {
	lo_byte := c.stackPull()
	hi_byte := c.stackPull()

	c.registers.PC = (word(hi_byte) << 8 | word(lo_byte)) + 1
	return 0
}

func sbc(c *cpu6502) int {
	val := word(c.fetch())

	if c.flags.D == 0 {
		// Binary mode
		// See http://www.righto.com/2012/12/the-6502-overflow-flag-explained.html
		// val is converted to ones complement and hence we can use ADC logic
		val = val ^ 0x00FF
		total := word(c.registers.A) + val + word(c.flags.C)

		// Overflow check
		// See http://www.righto.com/2012/12/the-6502-overflow-flag-explained.html
		c.flags.V = 0
		if ((val ^ total) & (word(c.registers.A) ^ total) & 0x80) != 0 {
			c.flags.V = 1
		}
		c.registers.A = byte(total)
		c.flags.C = 0

		if total > 0xFF { c.flags.C = 1 }
	} else {
		// Decimal mode
		a := int(c.registers.A)
		b := int(val)
		temp := (a & 0x0F) - (b & 0x0F) + int(c.flags.C) - 1

		if temp < 0 {
			temp = ((temp - 0x06) & 0x0F) - 0x10
		}

		a = (a & 0xF0) - (b & 0xF0) + temp

		if a < 0 {
			a = a - 0x60
		}

		val = val ^ 0x00FF
		var total word = word(c.registers.A) + val + word(c.flags.C)

		c.flags.V = 0
		if ((val ^ total) & (word(c.registers.A) ^ total) & 0x80) != 0 {
			c.flags.V = 1
		}

		c.registers.A = byte(a)
		c.flags.C = 1
		if a < 0 { c.flags.C = 0}
	}

	c.setNZFlag(c.registers.A)
	return 0
}

func sec(c *cpu6502) int {
	c.flags.C = 1
	return 0
}

func sed(c *cpu6502) int {
	c.flags.D = 1
	return 0
}

func sei(c *cpu6502) int {
	c.flags.I = 1
	return 0
}

func sta(c *cpu6502) int {
	c.Memory[c.absoluteAddr] = c.registers.A
	return 0
}
func stx(c *cpu6502) int {
	c.Memory[c.absoluteAddr] = c.registers.X
	return 0
}

func sty(c *cpu6502) int {
	c.Memory[c.absoluteAddr] = c.registers.Y
	return 0
}

func tax(c *cpu6502) int {
	c.registers.X = c.registers.A
	c.setNZFlag(c.registers.X)
	return 0
}

func tay(c *cpu6502) int {
	c.registers.Y = c.registers.A
	c.setNZFlag(c.registers.Y)
	return 0
}

func tsx(c *cpu6502) int {
	c.registers.X = c.registers.SP
	c.setNZFlag(c.registers.X)
	return 0
}

func txa(c *cpu6502) int {
	c.registers.A = c.registers.X
	c.setNZFlag(c.registers.A)
	return 0
}

func txs(c *cpu6502) int {
	c.registers.SP = c.registers.X
	return 0
}

func tya(c *cpu6502) int {
	c.registers.A = c.registers.Y
	c.setNZFlag(c.registers.A)
	return 0
}