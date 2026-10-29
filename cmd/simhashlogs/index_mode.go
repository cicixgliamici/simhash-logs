package main

import "fmt"

const (
	indexBrute   = "brute"
	indexLSH     = "lsh"
	indexPaper   = "paper"
	indexPaperV1 = "paper-v1"
	indexManku   = "manku"
)

func validateIndexMode(mode string) error {
	switch mode {
	case indexBrute, indexLSH, indexPaper, indexPaperV1, indexManku:
		return nil
	default:
		return fmt.Errorf("index must be one of brute, lsh, paper (or paper-v1), manku")
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
