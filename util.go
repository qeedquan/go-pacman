package main

import (
	"log"
	"math/rand"
)

func ek(err error) {
	if err != nil {
		log.Println(err)
	}
}

func ck(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func randn(a, b int) int {
	return a + rand.Intn(b-a)
}

func clamp(x, a, b int) int {
	if x < a {
		x = a
	} else if x > b {
		x = b
	}
	return x
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
