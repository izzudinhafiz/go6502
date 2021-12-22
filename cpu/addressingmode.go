package cpu6502

func accumulator(c *cpu6502) int {
	c.fetched = c.registers.A
	return 0
}

func implicit(c *cpu6502) int {
	return 0
}

func immediate(c *cpu6502) int {
	c.absoluteAddr = c.registers.PC
	c.registers.PC += 1

	return 0
}

func zeropage(c *cpu6502) int {
	c.absoluteAddr = word(c.fetchByte())
	return 0
}

func zeropagex(c *cpu6502) int {
	c.absoluteAddr = word(c.fetchByte() + c.registers.X)
	return 0
}

func zeropagey(c *cpu6502) int {
	c.absoluteAddr = word(c.fetchByte() + c.registers.Y)
	return 0
}

func absolute(c *cpu6502) int {
	c.absoluteAddr = c.fetchWord()
	return 0
}

func absolutex(c *cpu6502) int {
	addr := c.fetchWord()
	c.absoluteAddr = addr + word(c.registers.X)

	// Extra cycle if we cross page boundaries
	if c.absoluteAddr & 0xFF00 != addr & 0xFF00 {
		return 1
	}

	return 0
}

func absolutey(c *cpu6502) int {
	addr := c.fetchWord()
	c.absoluteAddr = addr + word(c.registers.Y)

	// Extra cycle if we cross page boundaries
	if c.absoluteAddr & 0xFF00 != addr & 0xFF00 {
		return 1
	}

	return 0
}

func indirect(c *cpu6502) int {
	addr := c.fetchWord()

	if addr & 0XFF == 0XFF {
		// Simulate a page boundary hardware bug
		// See https://www.youtube.com/watch?v=8XmxKPJDGU0
		addr &= 0XFF00
	}
	c.absoluteAddr = c.ReadWord(addr)

	return 0
}

func indirectx(c *cpu6502) int {
	pointer := c.fetchByte() + c.registers.X
	lo_byte := c.Memory[pointer & 0xFF]
	hi_byte := c.Memory[(pointer + 1) & 0xFF]
	c.absoluteAddr = word(hi_byte) << 8 | word(lo_byte)

	return 0
}

func indirecty(c *cpu6502) int {
	// See https://stackoverflow.com/questions/46262435/indirect-y-indexed-addressing-mode-in-mos-6502
	// Also https://www.c64-wiki.com/wiki/Indirect-indexed_addressing
	vector := c.fetchByte()
	pointer := c.ReadWord(word(vector))

	c.absoluteAddr = pointer + word(c.registers.Y)

	// Extra cycle if we cross page boundaries
	if c.absoluteAddr & 0xFF00 != pointer & 0xFF00 {
		return 1
	}

	return 0
}

func relative(c *cpu6502) int {
	c.relativeAddr = word(c.fetchByte())

	if c.relativeAddr & 0x80 > 0 {
		c.relativeAddr |= 0xFF00
	}
	return 0
}