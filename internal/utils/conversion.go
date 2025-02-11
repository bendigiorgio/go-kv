package utils

import "strconv"

func BytesToMb(b int) float32 {
	return float32(b) / 1024 / 1024
}

func StringToInt(str string, defaultNo int) int {
	if str == "" {
		return defaultNo
	}
	i, err := strconv.Atoi(str)
	if err != nil {
		return defaultNo
	}

	return i
}
