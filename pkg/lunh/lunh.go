package lunh

// CheckNumber checks the number for validity using the Lunh algorithm.
// https://en.wikipedia.org/wiki/Luhn_algorithm
func CheckNumber(number string) bool {
	var n, d, i, m int

	for i = len(number) - 1; i >= 0; i-- {
		c := number[i]

		switch {
		case c == ' ':
			continue
		case c >= '0' && c <= '9':
			m = int(c - '0')
			if d%2 == 1 {
				m <<= 1
			}
			if m > 9 {
				m -= 9
			}
			n += m
			d++
		default:
			return false
		}
	}

	return d > 1 && n%10 == 0
}
