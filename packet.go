package main

var (
	bitMasks = []int{0, 1, 3, 7, 15, 31, 63, 127, 255, 511, 1023, 2047, 4095, 8191,
		16383, 32767, 65535, 0x1ffff, 0x3ffff, 0x7ffff, 0xfffff, 0x1fffff, 0x3fffff,
		0x7fffff, 0xffffff, 0x1ffffff, 0x3ffffff, 0x7ffffff, 0xfffffff, 0x1fffffff, 0x3fffffff,
		0x7fffffff, -1,
	}
)

type RSCPacket struct {
	buf []byte
	i   int
}

func (p *RSCPacket) inc() int {
	x := p.i
	p.i++
	return x
}

func (p *RSCPacket) hasReadableBytes() bool {
	return p.i < len(p.buf)
}

func (p *RSCPacket) readByte() byte {
	return p.buf[p.inc()]
}

func (p *RSCPacket) readByteAsInt() int {
	return int(p.buf[p.inc()])
}

func (p *RSCPacket) readShort() int {
	return (int(p.buf[p.inc()]) << 8) + int(p.buf[p.inc()])
}

func (p *RSCPacket) readInt() int {
	return (int(p.buf[p.inc()]) << 24) + (int(p.buf[p.inc()]) << 16) + (int(p.buf[p.inc()]) << 8) + int(p.buf[p.inc()])
}

func (p *RSCPacket) readSmart08_16() int {
	b1 := p.peekByte()
	if b1 < 128 {
		return p.readByteAsInt()
	} else {
		return p.readShort() - 0x8000
	}
}

func (p *RSCPacket) peekByte() byte {
	return p.buf[p.i]
}

func (p *RSCPacket) writeByte(b byte) {
	p.buf[p.inc()] = b
}

func (p *RSCPacket) writeShort(s int) {
	p.buf[p.inc()] = byte(s >> 8)
	p.buf[p.inc()] = byte(s)
}

func (p *RSCPacket) writeInt(s int) {
	p.buf[p.inc()] = byte(s >> 24)
	p.buf[p.inc()] = byte(s >> 16)
	p.buf[p.inc()] = byte(s >> 8)
	p.buf[p.inc()] = byte(s)
}

func (p *RSCPacket) writeLong(s int64) {
	p.buf[p.inc()] = byte(s >> 56)
	p.buf[p.inc()] = byte(s >> 48)
	p.buf[p.inc()] = byte(s >> 40)
	p.buf[p.inc()] = byte(s >> 32)
	p.buf[p.inc()] = byte(s >> 24)
	p.buf[p.inc()] = byte(s >> 16)
	p.buf[p.inc()] = byte(s >> 8)
	p.buf[p.inc()] = byte(s)
}

func (p *RSCPacket) writeSmart08_16(val int) {
	if val >= 0 && val < 128 {
		p.writeByte(byte(val))
	} else {
		p.writeShort(0x8000 + val)
	}
}

func (p *client) writeHuffman(s string) {
	p.encodeHuffman(s)
	p.writeBytes(p.encodedChatBuffer[:p.encodedChatLength])
}

func (p *RSCPacket) writeBytes(bs []byte) {
	copy(p.buf[p.i:], bs)
	p.i += len(bs)
}

func (p *RSCPacket) createPacket(opcode byte) {
	p.i = 0
	p.writeShort(0)
	p.writeByte(opcode)
}

func (p *RSCPacket) finish() {
	length := p.i - 2
	p.buf[0] = byte(length >> 8)
	p.buf[1] = byte(length)
}

func (p *RSCPacket) readBytes(n int) []byte {
	buf := p.buf[p.i : p.i+n]
	p.i += n
	return buf
}

func (p *RSCPacket) readString() string {
	startIdx := p.i
	for {
		if p.buf[p.i] == 10 {
			p.i++
			break
		}
		p.i++
	}
	return string(p.buf[startIdx : p.i-1])
}

func readBits(buffer []byte, start, length int) int {
	byteOffset := start >> 3
	bitOffset := 8 - (start & 7)
	bits := 0

	for ; length > bitOffset; bitOffset = 8 {
		bits += (int(buffer[byteOffset]) & bitMasks[bitOffset]) << (length - bitOffset)
		byteOffset++
		length -= bitOffset
	}

	if length == bitOffset {
		bits += int(buffer[byteOffset]) & bitMasks[bitOffset]
	} else {
		bits += (int(buffer[byteOffset]) >> (bitOffset - length)) & bitMasks[length]
	}

	return bits
}

func (p *RSCPacket) readHuffman() string {
	count := p.readSmart08_16()
	src := p.buf[p.i:]
	dest := make([]byte, count)
	decodeHuffman(src, dest, 0, 0, count)
	return string(dest)
}
