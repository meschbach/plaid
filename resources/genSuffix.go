package resources

import "math/rand"

func GenSuffix(maxLength int) string {
	length := int(rand.Int31n(int32(maxLength)) + 3)
	suffix := ""
	for len(suffix) < length {
		c := rand.Int31n(26)
		char := int('a') + int(c)
		suffix = suffix + string(rune(char))
	}
	return suffix
}
