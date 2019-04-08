package main

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// Number of Miller-Rabin tests
const c = 25

// A random function which generates a random big number, using crypto/rand
// crypto secure Golang library.
func RandomDev(bitLen int) (*big.Int, error) {
	randNum := big.NewInt(0)
	if bitLen <= 0 {
		return randNum, fmt.Errorf("bitlen should be greater than 0, but it is %d", bitLen)
	}
	byteLen := bitLen / 8
	byteRem := bitLen % 8
	if byteRem != 0 {
		byteLen++
	}
	rawRand := make([]byte, byteLen)

	for found := false; !found; found = randNum.BitLen() == bitLen {
		_, err := rand.Read(rawRand)
		if err != nil {
			return randNum, err
		}
		randNum.SetBytes(rawRand)
		// set MSBs to 0 to get a bitLen equal to bitLen param.
		var bit int
		for bit = randNum.BitLen() - 1; bit >= bitLen; bit-- {
			randNum.SetBit(randNum, bit, 0)
		}
		// Set bit number (bitLen-1) to 1
		randNum.SetBit(randNum, bit, 1)
	}
	return randNum, nil
}

// Returns the next prime number based on a specific number, checking for its primality
// using ProbablyPrime function.

func nextPrime(num *big.Int, n int) *big.Int {

	// Possible prime should be odd
	num.SetBit(num,0, 1)
	for ;!num.ProbablyPrime(c);  {
		// I add two to the number to obtain another odd number
		num.Add(num, big.NewInt(2))
	}
	return num
}

// Returns a random prime of length bitLen, using a given random function randFn.
func randomPrime(bitLen int, randFn func(int) (*big.Int, error)) (*big.Int, error) {

	num := new(big.Int)
	var err error

	if randFn == nil {
		return big.NewInt(0), fmt.Errorf("random function cannot be nil")
	}
	if bitLen <= 0 {
		return big.NewInt(0), fmt.Errorf("bit length must be positive")
	}

	// Obtain a random number of length bitLen
	for true {
		num, err = randFn(bitLen)
		if err != nil {
			return num, err
		}
		num = nextPrime(num, c)

		// my next random number is too high
		if num.BitLen() == bitLen {
			break
		}
	}

	if num.BitLen() != bitLen {
		return big.NewInt(0), fmt.Errorf("random number returned should have length %d, but its length is %d", bitLen, num.BitLen())
	}

	if !num.ProbablyPrime(c) {
		return big.NewInt(0), fmt.Errorf("random number returned is not prime")
	}
	return num, nil
}

// Fast Safe Prime Generation.
// If it finds a prime, it tries the next probably safe prime or the previous one.
func GenerateSafePrime(bitLen int, randFn func(int) (*big.Int, error)) (*big.Int, error) {
	if randFn == nil {
		return big.NewInt(0), fmt.Errorf("random function cannot be nil")
	}

	q := new(big.Int)
	r := new(big.Int)


	for true {
		p, err := randomPrime(bitLen, randFn)
		if err != nil {
			return big.NewInt(0), err
		}
		// q is the first candidate = (p - 1) / 2
		q.Quo(big.NewInt(0).Sub(p, big.NewInt(1)), big.NewInt(2))

		// r is the second candidate = (p + 1) * 2
		r.Mul(big.NewInt(0).Add(p, big.NewInt(1)), big.NewInt(2))

		if r.ProbablyPrime(c) {
			return r, nil
		}
		if q.ProbablyPrime(c) {
			return p, nil
		}
	}

	return big.NewInt(0), fmt.Errorf("should never be here")
}
