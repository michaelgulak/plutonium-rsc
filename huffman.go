package main

// This code was taken from the OpenRSC server, which was refactored from the RSC client.

var (
	specialCharacters = []rune{'\u20ac', '?', '\u201a', '\u0192', '\u201e', '\u2026', '\u2020', '\u2021', '\u02c6',
		'\u2030', '\u0160', '\u2039', '\u0152', '?', '\u017d', '?', '?', '\u2018', '\u2019', '\u201c',
		'\u201d', '\u2022', '\u2013', '\u2014', '\u02dc', '\u2122', '\u0161', '\u203a', '\u0153', '?',
		'\u017e', '\u0178'}
	block               []int32
	dictionary          []int32
	specialCharacterMap map[rune]rune
	init0               = []int8{22, 22, 22, 22, 22, 22, 21, 22,
		22, 20, 22, 22, 22, 21, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 3, 8, 22,
		16, 22, 16, 17, 7, 13, 13, 13, 16, 7, 10, 6, 16, 10, 11, 12, 12, 12, 12, 13, 13, 14, 14, 11, 14, 19, 15, 17,
		8, 11, 9, 10, 10, 10, 10, 11, 10, 9, 7, 12, 11, 10, 10, 9, 10, 10, 12, 10, 9, 8, 12, 12, 9, 14, 8, 12, 17,
		16, 17, 22, 13, 21, 4, 7, 6, 5, 3, 6, 6, 5, 4, 10, 7, 5, 6, 4, 4, 6, 10, 5, 4, 4, 5, 7, 6, 10, 6, 10, 22,
		19, 22, 14, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22,
		22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22,
		22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22,
		22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22,
		22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 21, 22, 21, 22, 22, 22, 21, 22, 22}
)

func init() {
	dictionary = make([]int32, 8)
	specialCharacterMap = map[rune]rune{}
	block = make([]int32, len(init0))
	for i := 0; i < len(specialCharacters); i++ {
		specialCharacterMap[specialCharacters[i]] = rune(i - 128)
	}
	blockBuilder := make([]int32, 33)
	dictIndexTemp := int32(0)
	for initPos := int32(0); initPos < int32(len(init0)); initPos++ {
		initValue := init0[initPos]
		builderBitSelector := int32(1 << (32 - initValue))
		builderValue := blockBuilder[initValue]
		block[initPos] = builderValue
		var builderValueBit int32
		if builderValue&builderBitSelector == 0 {
			builderValueBit = builderValue | builderBitSelector
			for initValueCounter := initValue - 1; initValueCounter > 0; initValueCounter-- {
				builderValue2 := blockBuilder[initValueCounter]
				if builderValue != builderValue2 {
					break
				}
				builderValue2BitSelector := int32(1 << (32 - initValueCounter))

				if builderValue2&builderValue2BitSelector == 0 {
					blockBuilder[initValueCounter] = builderValue2BitSelector | builderValue2
				} else {
					blockBuilder[initValueCounter] = blockBuilder[initValueCounter-1]
					break
				}
			}
		} else {
			builderValueBit = blockBuilder[initValue+-1]
		}
		blockBuilder[initValue] = builderValueBit
		for initValueCounter := initValue + 1; initValueCounter <= 32; initValueCounter++ {
			if builderValue == blockBuilder[initValueCounter] {
				blockBuilder[initValueCounter] = builderValueBit
			}
		}
		dictIndex := int32(0)
		for initValueCounter := 0; initValueCounter < int(initValue); initValueCounter++ {
			builderBitSelector2 := int32(uint32(0x80000000) >> initValueCounter)
			if builderValue&builderBitSelector2 == 0 {
				dictIndex++
			} else {
				if dictionary[dictIndex] == 0 {
					dictionary[dictIndex] = dictIndexTemp
				}
				dictIndex = dictionary[dictIndex]
			}
			if int32(len(dictionary)) <= dictIndex {
				newDict := make([]int32, len(dictionary)*2)
				copy(newDict[:len(dictionary)], dictionary)
				dictionary = newDict
			}
		}
		dictionary[dictIndex] = ^initPos
		if dictIndex >= dictIndexTemp {
			dictIndexTemp = dictIndex + 1
		}
	}
}

func (p *client) convertMessageToBytes(s string) {
	for i, c := range s {
		if ((0 >= c) || ('\u0080' <= c)) && (('\u00a0' > c) || ('\u00ff' < c)) {
			if c0, ok := specialCharacterMap[c]; !ok {
				p.chatBuffer[i] = '?'
			} else {
				p.chatBuffer[i] = byte(c0)
			}
		} else {
			p.chatBuffer[i] = byte(c)
		}
	}
}

func (c *client) encodeHuffman(message string) {
	c.convertMessageToBytes(message)
	encodedByte := 0
	outputBitOffset := 0
	for messageIndex := 0; len(message) > messageIndex; messageIndex++ {
		messageCharacter := c.chatBuffer[messageIndex]
		blockValue := block[messageCharacter]
		initValue := init0[messageCharacter]
		outputByteOffset := uint32(outputBitOffset) >> 3
		blockShifter := 0x7 & outputBitOffset
		encodedByte &= int(-blockShifter >> 31)
		outputByteOffset2 := outputByteOffset + uint32(blockShifter+int(initValue)-1)>>3
		outputBitOffset += int(initValue)
		blockShifter += 24
		encodedByte |= int(uint32(blockValue) >> blockShifter)
		c.encodedChatBuffer[outputByteOffset] = byte(encodedByte)
		if outputByteOffset2 > outputByteOffset {
			outputByteOffset++
			blockShifter -= 8
			encodedByte = int(uint32(blockValue) >> blockShifter)
			c.encodedChatBuffer[outputByteOffset] = byte(encodedByte)
			if outputByteOffset < outputByteOffset2 {
				outputByteOffset++
				blockShifter -= 8
				encodedByte = int(uint32(blockValue) >> blockShifter)
				c.encodedChatBuffer[outputByteOffset] = byte(encodedByte)
				if outputByteOffset2 > outputByteOffset {
					outputByteOffset++
					blockShifter -= 8
					encodedByte = int(uint32(blockValue) >> blockShifter)
					c.encodedChatBuffer[outputByteOffset] = byte(encodedByte)
					if outputByteOffset2 > outputByteOffset {
						blockShifter -= 8
						outputByteOffset++
						encodedByte = int(blockValue << -blockShifter)
						c.encodedChatBuffer[outputByteOffset] = byte(encodedByte)
					}
				}
			}
		}
	}
	c.decodedChatLength = len(message)
	c.encodedChatLength = int(uint32(outputBitOffset+7) >> 3)
}

func decodeHuffman(src []byte, dest []byte, destOffset int, srcOffset int, count int) int {
	if count == 0 {
		return 0
	} else {
		var7 := int32(0)
		count += destOffset
		var8 := srcOffset
		for {
			var9 := int8(src[var8])
			if var9 >= 0 {
				var7++
			} else {
				var7 = dictionary[var7]
			}
			var var10 = dictionary[var7]
			if var10 < 0 {
				dest[destOffset] = byte(^var10)
				destOffset++
				if destOffset >= count {
					break
				}
				var7 = 0
			}
			if 64&var9 != 0 {
				var7 = dictionary[var7]
			} else {
				var7++
			}
			var10 = dictionary[var7]
			if var10 < 0 {
				dest[destOffset] = byte(^var10)
				destOffset++
				if count <= destOffset {
					break
				}
				var7 = 0
			}
			if var9&32 == 0 {
				var7++
			} else {
				var7 = dictionary[var7]
			}
			var10 = dictionary[var7]
			if var10 < 0 {
				dest[destOffset] = byte(^var10)
				destOffset++
				if count <= destOffset {
					break
				}
				var7 = 0
			}
			if 16&var9 != 0 {
				var7 = dictionary[var7]
			} else {
				var7++
			}
			var10 = dictionary[var7]
			if var10 < 0 {
				dest[destOffset] = byte(^var10)
				destOffset++
				if destOffset >= count {
					break
				}
				var7 = 0
			}
			if var9&8 != 0 {
				var7 = dictionary[var7]
			} else {
				var7++
			}
			var10 = dictionary[var7]
			if var10 < 0 {
				dest[destOffset] = byte(^var10)
				destOffset++
				if destOffset >= count {
					break
				}
				var7 = 0
			}
			if 4&var9 != 0 {
				var7 = dictionary[var7]
			} else {
				var7++
			}
			var10 = dictionary[var7]
			if var10 < 0 {
				dest[destOffset] = byte(^var10)
				destOffset++
				if destOffset >= count {
					break
				}
				var7 = 0
			}
			if 2&var9 == 0 {
				var7++
			} else {
				var7 = dictionary[var7]
			}
			var10 = dictionary[var7]
			if var10 < 0 {
				dest[destOffset] = byte(^var10)
				destOffset++
				if destOffset >= count {
					break
				}
				var7 = 0
			}
			if 1&var9 != 0 {
				var7 = dictionary[var7]
			} else {
				var7++
			}
			var10 = dictionary[var7]
			if var10 < 0 {
				dest[destOffset] = byte(^var10)
				destOffset++
				if destOffset >= count {
					break
				}
				var7 = 0
			}
			var8++
		}
		return 1 - srcOffset + var8
	}
}
