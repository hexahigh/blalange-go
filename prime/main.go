package main

import (
	"fmt"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Usage:", os.Args[0], "N")
		os.Exit(1)
	}

	N, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: failed to parse N as an integer; is it a number?")
		os.Exit(1)
	}

	// Create a slice to mark numbers as prime
	primes := make([]bool, N+1)
	for i := 2; i <= N; i++ {
		primes[i] = true
	}

	// Sieve of Eratosthenes
	for p := 2; p*p <= N; p++ {
		if primes[p] {
			for i := p * p; i <= N; i += p {
				primes[i] = false
			}
		}
	}

	// Print the prime numbers
	for p := 2; p <= N; p++ {
		if primes[p] {
			fmt.Println(p)
		}
	}
}
