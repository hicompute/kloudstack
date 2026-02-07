package helper

import (
	"math/big"
	"strings"
)

func StringToBigInt(s string) *big.Int {
	if s == "" {
		return big.NewInt(0)
	}
	bi, ok := new(big.Int).SetString(s, 10)
	if !ok {
		return big.NewInt(0)
	}
	return bi
}

func ExtractVMName(podName string) string {
	prefix := "virt-launcher-"

	// Check if it starts with the correct prefix
	if !strings.HasPrefix(podName, prefix) {
		return ""
	}

	// Remove the prefix
	nameWithoutPrefix := strings.TrimPrefix(podName, prefix)

	// Find the last dash to separate VM name from random suffix
	lastDashIndex := strings.LastIndex(nameWithoutPrefix, "-")
	if lastDashIndex == -1 {
		// No dash found after prefix means no random suffix
		return ""
	}

	// Extract everything before the last dash (the VM name)
	vmName := nameWithoutPrefix[:lastDashIndex]

	// Return empty if VM name would be empty
	if vmName == "" {
		return ""
	}

	return vmName
}
