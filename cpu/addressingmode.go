package cpu6502

func accumulator(c *Cpu6502) int {
	c.Fetched = c.Registers.A
	return 0
}

func implicit(c *Cpu6502) int {
	return 0
}

func immediate(c *Cpu6502) int {
	c.AbsoluteAddr = c.Registers.PC
	c.Registers.PC += 1

	return 0
}

func zeropage(c *Cpu6502) int {
	c.AbsoluteAddr = word(c.fetchByte())
	return 0
}

func zeropagex(c *Cpu6502) int {
	c.AbsoluteAddr = word(c.fetchByte() + c.Registers.X)
	return 0
}

func zeropagey(c *Cpu6502) int {
	c.AbsoluteAddr = word(c.fetchByte() + c.Registers.Y)
	return 0
}

func absolute(c *Cpu6502) int {
	c.AbsoluteAddr = c.fetchWord()
	return 0
}

func absolutex(c *Cpu6502) int {
	addr := c.fetchWord()
	c.AbsoluteAddr = addr + word(c.Registers.X)

	// Extra cycle if we cross page boundaries
	if c.AbsoluteAddr & 0xFF00 != addr & 0xFF00 {
		return 1
	}

	return 0
}

func absolutey(c *Cpu6502) int {
	addr := c.fetchWord()
	c.AbsoluteAddr = addr + word(c.Registers.Y)

	// Extra cycle if we cross page boundaries
	if c.AbsoluteAddr & 0xFF00 != addr & 0xFF00 {
		return 1
	}

	return 0
}

func indirect(c *Cpu6502) int {
	addr := c.fetchWord()

	if addr & 0XFF == 0XFF {
		// Simulate a page boundary hardware bug
		// See https://www.youtube.com/watch?v=8XmxKPJDGU0
		addr &= 0XFF00
	}
	c.AbsoluteAddr = c.ReadWord(addr)

	return 0
}

func indirectx(c *Cpu6502) int {
	pointer := c.fetchByte() + c.Registers.X
	lo_byte := c.Memory[pointer & 0xFF]
	hi_byte := c.Memory[(pointer + 1) & 0xFF]
	c.AbsoluteAddr = word(hi_byte) << 8 | word(lo_byte)

	return 0
}

func indirecty(c *Cpu6502) int {
	// See https://stackoverflow.com/questions/46262435/indirect-y-indexed-addressing-mode-in-mos-6502
	// Also https://www.c64-wiki.com/wiki/Indirect-indexed_addressing
	vector := c.fetchByte()
	pointer := c.ReadWord(word(vector))

	c.AbsoluteAddr = pointer + word(c.Registers.Y)

	// Extra cycle if we cross page boundaries
	if c.AbsoluteAddr & 0xFF00 != pointer & 0xFF00 {
		return 1
	}

	return 0
}

func relative(c *Cpu6502) int {
	c.RelativeAddr = word(c.fetchByte())

	if c.RelativeAddr & 0x80 > 0 {
		c.RelativeAddr |= 0xFF00
	}
	return 0
}