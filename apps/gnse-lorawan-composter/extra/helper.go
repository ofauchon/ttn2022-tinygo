package extra

import "crypto/rand"

// byteToHex return string hex representation of byte
func ByteToHex(b byte) string {
	bb := (b >> 4) & 0x0F
	ret := ""
	if bb < 10 {
		ret += string(rune('0' + bb))
	} else {
		ret += string(rune('A' + (bb - 10)))
	}

	bb = (b) & 0xF
	if bb < 10 {
		ret += string(rune('0' + bb))
	} else {
		ret += string(rune('A' + (bb - 10)))
	}
	return ret
}

// BytesToHexString converts byte slice to hex string representation
func BytesToHexString(data []byte) string {
	s := ""
	for i := 0; i < len(data); i++ {
		s += ByteToHex(data[i])
	}
	return s
}

// GetRand16 returns 2 random bytes
func GetRand16() [2]uint8 {
	randomBytes := make([]byte, 2)
	_, err := rand.Read(randomBytes)
	if err == nil {
		r16 := [2]uint8{randomBytes[0], randomBytes[1]}
		return r16
	} else {
		println("RAND ERROR ... Abort")
		for {
		}
	}

}
