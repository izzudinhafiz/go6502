package cpu6502

func twos_comp(val int) int {
	if val&(1<<7) > 0 {
		val = val - 256
	}

	return val
}

func IRQ(c *Cpu6502) int {
	if c.Flags.I == 0 {
		c.stackPush(byte(c.Registers.PC >> 8))
		c.stackPush(byte(c.Registers.PC))

		c.Flags.I = 1
		c.stackPush(c.getStatusFlagsByte("interrupt"))

		c.Registers.PC = (word(c.Memory[0xFFFF]) << 8) | word(c.Memory[0xFFFE])

		return 7
	}
	return 0
}

func NMI(c *Cpu6502) int {
	c.stackPush(byte(c.Registers.PC >> 8))
	c.stackPush(byte(c.Registers.PC))

	c.Flags.I = 1
	c.stackPush(c.getStatusFlagsByte("interrupt"))
	c.Registers.PC = (word(c.Memory[0xFFFB]) << 8) | word(c.Memory[0xFFFA])
	return 8
}

func nop(c *Cpu6502) int {
	return 0
}

func adc(c *Cpu6502) int {
	// Add with carry operation
	// Decimal mode: N, V, Z flags are invalid
	val := c.fetchByte()

	// Binary mode
	if c.Flags.D == 0 {
		total := word(c.Registers.A) + word(val) + word(c.Flags.C)

		// Overflow check
		// See http://www.righto.com/2012/12/the-6502-overflow-flag-explained.html
		if ((word(val) ^ total) & (word(c.Registers.A) ^ total) & 0x80) != 0 {
			c.Flags.V = 1
		} else {
			c.Flags.V = 0
		}

		if total > 0xFF {
			c.Flags.C = 1
		} else {
			c.Flags.C = 0
		}

		c.Registers.A = byte(total)
		c.setNZFlag(c.Registers.A)

		return 0
	} else {
		// Decimal mode
		a := int(c.Registers.A)
		b := int(val)
		temp := (a & 0x0F) + (b & 0x0F) + int(c.Flags.C) // Add 1s place of decimal

		if temp >= 0x0A { // If bigger than 9
			temp = ((temp + 0x06) & 0x0F) + 0x10 // Add 6 to skip 10 -> 15, keep the 1s place and carry to 10s place
		}

		a = (a & 0xF0) + (b & 0xF0) + temp // Add the 10s place in decimal

		if (a&0xFF)&(1<<7) > 0 { // N set if 7th bit is set
			c.Flags.N = 1
		} else {
			c.Flags.N = 0
		}

		if -128 <= twos_comp(a&0xFF) && twos_comp(a&0xFF) <= 127 {
			c.Flags.V = 0
		} else {
			c.Flags.V = 1
		}

		if a >= 0xA0 { // If bigger than 100
			a = a + 0x60 // skip 1xx -> 5xx, keeps the 10s and 1s place
		}

		c.Registers.A = byte(a)
		if a >= 0x100 {
			c.Flags.C = 1
		} else {
			c.Flags.C = 0
		}

		return 0
	}
}

func and(c *Cpu6502) int {
	val := c.fetch()
	c.Registers.A = val & c.Registers.A
	c.setNZFlag(c.Registers.A)

	return 0
}

func asl(c *Cpu6502) int {
	var val word
	if c.Opcode.AddressingMode == ADR_ACCUMULATOR {
		val = word(c.Registers.A) << 1
		c.Registers.A = byte(val)
	} else {
		val = word(c.fetch()) << 1
		c.Memory[c.AbsoluteAddr] = byte(val)
	}

	c.Flags.C = 0
	if val > 255 {
		c.Flags.C = 1
	}
	c.setNZFlag(byte(val))
	return 0
}

func bcc(c *Cpu6502) int {
	if c.Flags.C == 0 {
		cycles := 1
		c.AbsoluteAddr = (c.Registers.PC + c.RelativeAddr)

		// If branch to new page, add a cycle
		if c.AbsoluteAddr&0xFF00 != c.Registers.PC&0xFF00 {
			cycles += 1
		}
		c.Registers.PC = c.AbsoluteAddr

		return cycles
	}
	return 0
}

func bcs(c *Cpu6502) int {
	if c.Flags.C == 1 {
		cycles := 1
		c.AbsoluteAddr = (c.Registers.PC + c.RelativeAddr)

		// If branch to new page, add a cycle
		if c.AbsoluteAddr&0xFF00 != c.Registers.PC&0xFF00 {
			cycles += 1
		}
		c.Registers.PC = c.AbsoluteAddr

		return cycles
	}
	return 0
}

func beq(c *Cpu6502) int {
	if c.Flags.Z == 1 {
		cycles := 1
		c.AbsoluteAddr = (c.Registers.PC + c.RelativeAddr)

		// If branch to new page, add a cycle
		if c.AbsoluteAddr&0xFF00 != c.Registers.PC&0xFF00 {
			cycles += 1
		}
		c.Registers.PC = c.AbsoluteAddr

		return cycles
	}
	return 0
}

func bit(c *Cpu6502) int {
	val := c.fetch()
	c.Flags.Z = 0
	c.Flags.N = 0
	c.Flags.V = 0

	if c.Registers.A&val == 0 {
		c.Flags.Z = 1
	}
	if val&(1<<7) > 0 {
		c.Flags.N = 1
	}
	if val&(1<<6) > 0 {
		c.Flags.V = 1
	}

	return 0
}

func bmi(c *Cpu6502) int {
	if c.Flags.N == 1 {
		cycles := 1
		c.AbsoluteAddr = (c.Registers.PC + c.RelativeAddr)

		// If branch to new page, add a cycle
		if c.AbsoluteAddr&0xFF00 != c.Registers.PC&0xFF00 {
			cycles += 1
		}
		c.Registers.PC = c.AbsoluteAddr

		return cycles
	}
	return 0
}

func bne(c *Cpu6502) int {
	if c.Flags.Z == 0 {
		cycles := 1
		c.AbsoluteAddr = (c.Registers.PC + c.RelativeAddr)

		// If branch to new page, add a cycle
		if c.AbsoluteAddr&0xFF00 != c.Registers.PC&0xFF00 {
			cycles += 1
		}
		c.Registers.PC = c.AbsoluteAddr

		return cycles
	}
	return 0
}

func bpl(c *Cpu6502) int {
	if c.Flags.N == 0 {
		cycles := 1
		c.AbsoluteAddr = (c.Registers.PC + c.RelativeAddr)

		// If branch to new page, add a cycle
		if c.AbsoluteAddr&0xFF00 != c.Registers.PC&0xFF00 {
			cycles += 1
		}
		c.Registers.PC = c.AbsoluteAddr

		return cycles
	}
	return 0
}

func brk(c *Cpu6502) int {
	c.Registers.PC += 1
	c.stackPush(byte(c.Registers.PC >> 8)) // write high byte
	c.stackPush(byte(c.Registers.PC))      // write low byte

	// On a BRK instruction, the CPU does the same as in the IRQ case,
	// but sets bit #4 (B flag) IN THE COPY of the status register that is saved on the stack.
	c.stackPush(c.getStatusFlagsByte("instruction"))
	c.Flags.I = 1

	c.Registers.PC = word(c.Memory[0xFFFF])<<8 | word(c.Memory[0xFFFE])

	return 0
}

func bvc(c *Cpu6502) int {
	if c.Flags.V == 0 {
		cycles := 1
		c.AbsoluteAddr = (c.Registers.PC + c.RelativeAddr)

		// If branch to new page, add a cycle
		if c.AbsoluteAddr&0xFF00 != c.Registers.PC&0xFF00 {
			cycles += 1
		}
		c.Registers.PC = c.AbsoluteAddr

		return cycles
	}
	return 0
}

func bvs(c *Cpu6502) int {
	if c.Flags.V == 1 {
		cycles := 1
		c.AbsoluteAddr = (c.Registers.PC + c.RelativeAddr)

		// If branch to new page, add a cycle
		if c.AbsoluteAddr&0xFF00 != c.Registers.PC&0xFF00 {
			cycles += 1
		}
		c.Registers.PC = c.AbsoluteAddr

		return cycles
	}
	return 0
}

func clc(c *Cpu6502) int {
	c.Flags.C = 0
	return 0
}

func cld(c *Cpu6502) int {
	c.Flags.D = 0
	return 0
}

func cli(c *Cpu6502) int {
	c.Flags.I = 0
	return 0
}

func clv(c *Cpu6502) int {
	c.Flags.V = 0
	return 0
}

func cmp(c *Cpu6502) int {
	val := c.fetch()

	c.Flags.C = 0
	if c.Registers.A >= val {
		c.Flags.C = 1
	}
	c.setNZFlag(c.Registers.A - val)
	return 0
}

func cpx(c *Cpu6502) int {
	val := c.fetch()

	c.Flags.C = 0
	if c.Registers.X >= val {
		c.Flags.C = 1
	}
	c.setNZFlag(c.Registers.X - val)
	return 0
}

func cpy(c *Cpu6502) int {
	val := c.fetch()

	c.Flags.C = 0
	if c.Registers.Y >= val {
		c.Flags.C = 1
	}
	c.setNZFlag(c.Registers.Y - val)
	return 0
}

func dec(c *Cpu6502) int {
	val := c.fetch() - 1
	c.Memory[c.AbsoluteAddr] = val
	c.setNZFlag(val)
	return 0
}

func dex(c *Cpu6502) int {
	c.Registers.X -= 1
	c.setNZFlag(c.Registers.X)
	return 0
}

func dey(c *Cpu6502) int {
	c.Registers.Y -= 1
	c.setNZFlag(c.Registers.Y)
	return 0
}

func eor(c *Cpu6502) int {
	c.Registers.A = c.Registers.A ^ c.fetch()
	c.setNZFlag(c.Registers.A)
	return 0
}

func inc(c *Cpu6502) int {
	val := c.fetch() + 1
	c.Memory[c.AbsoluteAddr] = val
	c.setNZFlag(val)
	return 0
}

func inx(c *Cpu6502) int {
	c.Registers.X = c.Registers.X + 1
	c.setNZFlag(c.Registers.X)
	return 0
}

func iny(c *Cpu6502) int {
	c.Registers.Y = c.Registers.Y + 1
	c.setNZFlag(c.Registers.Y)
	return 0
}

func jmp(c *Cpu6502) int {
	c.Registers.PC = c.AbsoluteAddr
	return 0
}

func jsr(c *Cpu6502) int {
	c.Registers.PC -= 1
	c.stackPush(byte(c.Registers.PC >> 8)) // Write high byte
	c.stackPush(byte(c.Registers.PC))      // write low byte

	c.Registers.PC = c.AbsoluteAddr
	return 0
}

func lda(c *Cpu6502) int {
	c.Registers.A = c.fetch()
	c.setNZFlag(c.Registers.A)
	return 0
}

func ldx(c *Cpu6502) int {
	c.Registers.X = c.fetch()
	c.setNZFlag(c.Registers.X)
	return 0
}

func ldy(c *Cpu6502) int {
	c.Registers.Y = c.fetch()
	c.setNZFlag(c.Registers.Y)
	return 0
}

func lsr(c *Cpu6502) int {
	var val byte
	if c.Opcode.AddressingMode == ADR_ACCUMULATOR {
		val = c.Registers.A
	} else {
		val = c.fetch()
	}
	c.Memory[c.AbsoluteAddr] = val >> 1

	c.Flags.C = 0
	if val&0x01 > 0 {
		c.Flags.C = 1
	}
	c.setNZFlag(c.Flags.C)
	return 0
}

func ora(c *Cpu6502) int {
	c.Registers.A = c.Registers.A | c.fetch()
	c.setNZFlag(c.Registers.A)
	return 0
}

func pha(c *Cpu6502) int {
	c.stackPush(c.Registers.A)
	return 0
}

func pla(c *Cpu6502) int {
	c.Registers.A = c.stackPull()
	c.setNZFlag(c.Registers.A)
	return 0
}

func php(c *Cpu6502) int {
	c.stackPush(c.getStatusFlagsByte("instruction"))
	c.Flags.B = 0
	return 0
}

func plp(c *Cpu6502) int {
	c.setStatusFlags(c.stackPull())
	c.Flags.B = 0
	return 0
}

func rol(c *Cpu6502) int {
	var val word
	var temp byte
	if c.Opcode.AddressingMode == ADR_ACCUMULATOR {
		val = word(c.Registers.A)
		temp := byte(val << 1) | c.Flags.C
		c.Registers.A = temp
	} else {
		val = word(c.fetch())
		temp := byte(val << 1) | c.Flags.C
		c.Memory[c.AbsoluteAddr] = temp
	}

	c.Flags.C = 0
	if val & 0x01 > 0 { c.Flags.C = 1}
	c.setNZFlag(temp)
	return 0
}

func ror(c *Cpu6502) int {
	var val word
	var temp byte
	if c.Opcode.AddressingMode == ADR_ACCUMULATOR {
		val = word(c.Registers.A)
		temp := byte(val >> 1) | (c.Flags.C << 7)
		c.Registers.A = temp
	} else {
		val = word(c.fetch())
		temp := byte(val >> 1) | (c.Flags.C << 7)
		c.Memory[c.AbsoluteAddr] = temp
	}

	c.Flags.C = 0
	if val & 0x01 > 0 { c.Flags.C = 1}
	c.setNZFlag(temp)
	return 0
}

func rti(c *Cpu6502) int {
	c.setStatusFlags(c.stackPull())
	c.Flags.B = 0

	lo_byte := c.stackPull()
	hi_byte := c.stackPull()

	c.Registers.PC = word(hi_byte) << 8 | word(lo_byte)
	return 0
}

func rts(c *Cpu6502) int {
	lo_byte := c.stackPull()
	hi_byte := c.stackPull()

	c.Registers.PC = (word(hi_byte) << 8 | word(lo_byte)) + 1
	return 0
}

func sbc(c *Cpu6502) int {
	val := word(c.fetch())

	if c.Flags.D == 0 {
		// Binary mode
		// See http://www.righto.com/2012/12/the-6502-overflow-flag-explained.html
		// val is converted to ones complement and hence we can use ADC logic
		val = val ^ 0x00FF
		total := word(c.Registers.A) + val + word(c.Flags.C)

		// Overflow check
		// See http://www.righto.com/2012/12/the-6502-overflow-flag-explained.html
		c.Flags.V = 0
		if ((val ^ total) & (word(c.Registers.A) ^ total) & 0x80) != 0 {
			c.Flags.V = 1
		}
		c.Registers.A = byte(total)
		c.Flags.C = 0

		if total > 0xFF { c.Flags.C = 1 }
	} else {
		// Decimal mode
		a := int(c.Registers.A)
		b := int(val)
		temp := (a & 0x0F) - (b & 0x0F) + int(c.Flags.C) - 1

		if temp < 0 {
			temp = ((temp - 0x06) & 0x0F) - 0x10
		}

		a = (a & 0xF0) - (b & 0xF0) + temp

		if a < 0 {
			a = a - 0x60
		}

		val = val ^ 0x00FF
		var total word = word(c.Registers.A) + val + word(c.Flags.C)

		c.Flags.V = 0
		if ((val ^ total) & (word(c.Registers.A) ^ total) & 0x80) != 0 {
			c.Flags.V = 1
		}

		c.Registers.A = byte(a)
		c.Flags.C = 1
		if a < 0 { c.Flags.C = 0}
	}

	c.setNZFlag(c.Registers.A)
	return 0
}

func sec(c *Cpu6502) int {
	c.Flags.C = 1
	return 0
}

func sed(c *Cpu6502) int {
	c.Flags.D = 1
	return 0
}

func sei(c *Cpu6502) int {
	c.Flags.I = 1
	return 0
}

func sta(c *Cpu6502) int {
	c.Memory[c.AbsoluteAddr] = c.Registers.A
	return 0
}
func stx(c *Cpu6502) int {
	c.Memory[c.AbsoluteAddr] = c.Registers.X
	return 0
}

func sty(c *Cpu6502) int {
	c.Memory[c.AbsoluteAddr] = c.Registers.Y
	return 0
}

func tax(c *Cpu6502) int {
	c.Registers.X = c.Registers.A
	c.setNZFlag(c.Registers.X)
	return 0
}

func tay(c *Cpu6502) int {
	c.Registers.Y = c.Registers.A
	c.setNZFlag(c.Registers.Y)
	return 0
}

func tsx(c *Cpu6502) int {
	c.Registers.X = c.Registers.SP
	c.setNZFlag(c.Registers.X)
	return 0
}

func txa(c *Cpu6502) int {
	c.Registers.A = c.Registers.X
	c.setNZFlag(c.Registers.A)
	return 0
}

func txs(c *Cpu6502) int {
	c.Registers.SP = c.Registers.X
	return 0
}

func tya(c *Cpu6502) int {
	c.Registers.A = c.Registers.Y
	c.setNZFlag(c.Registers.A)
	return 0
}