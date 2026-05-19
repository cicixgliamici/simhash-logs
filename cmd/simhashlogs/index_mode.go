package main

import "fmt"

const (
	indexBrute = "brute"
	indexLSH   = "lsh"
	indexPaper = "paper"
)

func validateIndexMode(mode string) error {
	switch mode {
	case indexBrute, indexLSH, indexPaper:
		return nil
	default:
		return fmt.Errorf("index must be one of brute, lsh, paper")
	}
}

func defaultTables(k int) int {
	if k < 0 {
		return 1
	}
	if k+1 > 64 {
		return 64
	}
	return k + 1
}

func defaultWindow(k int) int {
	if k < 3 {
		return 3
	}
	return k
}
