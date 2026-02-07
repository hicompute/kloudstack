package netutils

import (
	"crypto/sha256"
	"fmt"
)

func GenerateVethMAC(input, prefix string) string {
	hash := sha256.Sum256([]byte(input))

	// XOR operations to better distribute the bits
	part1 := hash[1] ^ hash[7] ^ hash[13] ^ hash[19]
	part2 := hash[2] ^ hash[8] ^ hash[14] ^ hash[20]
	part3 := hash[3] ^ hash[9] ^ hash[15] ^ hash[21]
	part4 := hash[4] ^ hash[10] ^ hash[16] ^ hash[22]
	part5 := hash[5] ^ hash[11] ^ hash[17] ^ hash[23]

	return fmt.Sprintf(prefix+":%02x:%02x:%02x:%02x:%02x",
		part1, part2, part3, part4, part5)
}
