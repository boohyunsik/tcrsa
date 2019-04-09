package main

import (
	"crypto/sha256"
	"fmt"
	"math/big"
)

type SignatureShare struct {
	Xi []byte
	C  []byte
	Z  []byte
	Id uint16
}

type SignatureShares []*SignatureShare

type Signature []byte


func (sigShare SignatureShare) GetIndex() uint16 {
	return sigShare.Id - 1
}

func (sigShare SignatureShare) Verify(doc []byte, info *KeyMeta) bool {

	x := new(big.Int)
	xi := new(big.Int)
	z := new(big.Int)
	c := new(big.Int)
	c2 := new(big.Int)
	n := new(big.Int)
	e := new(big.Int)
	v := new(big.Int)
	u := new(big.Int)
	vki := new(big.Int)
	xTilde := new(big.Int)
	xi2 := new(big.Int)
	vPrime := new(big.Int)
	xPrime := new(big.Int)

	negC := new(big.Int)
	xiNeg2c := new(big.Int)
	aux := new(big.Int)

	x.SetBytes(doc)
	n.Set(info.PublicKey.N)
	e.SetUint64(uint64(info.PublicKey.E))
	v.SetBytes(info.VerificationKey.V)
	u.SetBytes(info.VerificationKey.U)
	vki.SetBytes(info.VerificationKey.I[sigShare.GetIndex()])

	xi.SetBytes(sigShare.Xi)
	z.SetBytes(sigShare.Z)
	c.SetBytes(sigShare.C)

	if big.Jacobi(x, n) == -1 {
		ue := new(big.Int).Exp(u, e, n)
		x.Mul(x, ue).Mod(x, n)
	}

	// x~ = x^4 % n
	xTilde.Exp(x, big.NewInt(4), n)

	// xi_2 = xi^2 % n
	xi2.Exp(xi, big.NewInt(2), n)

	// v' = v^z * v_i^(-c)
	negC.Neg(c)
	vPrime.Exp(vki, negC, n)
	aux.Exp(v, z, n)
	vPrime.Mul(vPrime, aux).Mod(vPrime, n)

	// x' = x~^z * x_i^(-2c)
	aux.Mul(negC, big.NewInt(2))
	xiNeg2c.Exp(xi, aux, n)

	aux.Exp(xTilde, z, n)
	xPrime.Mul(aux, xiNeg2c)
	xPrime.Mod(xPrime, n)

	// Hashing all the values
	sha := sha256.New()
	sha.Write(v.Bytes())
	sha.Write(u.Bytes())
	sha.Write(xTilde.Bytes())
	sha.Write(vki.Bytes())
	sha.Write(xi2.Bytes())
	sha.Write(vPrime.Bytes())
	sha.Write(xPrime.Bytes())

	hash := sha.Sum(nil)

	c2.SetBytes(hash)
	c2.Mod(c2, n)

	return c2.Cmp(c) == 0

}

// Joins the signatures of the document provided.
func (sigShares SignatureShares) Join(document []byte, info *KeyMeta) (Signature, error) {

	if document == nil {
		return []byte{}, fmt.Errorf("document is nil")
	}
	if info == nil {
		return []byte{}, fmt.Errorf("key metainfo is nil")
	}

	for i := 0; i < len(sigShares); i++ {
		if sigShares[i] == nil {
			return []byte{}, fmt.Errorf("signature share %d is nil", i)
		}
	}

	k := info.K
	if len(sigShares) < int(k) {
		return []byte{}, fmt.Errorf("insufficient number of signature shares. provided: %d, needed: %d", len(sigShares), k)
	}

	x := new(big.Int)
	n := new(big.Int)
	e := new(big.Int)
	u := new(big.Int)
	delta := new(big.Int)
	ePrime := new(big.Int)
	w := new(big.Int)
	si := new(big.Int)
	aux := new(big.Int)
	a := new(big.Int)
	b := new(big.Int)
	wa := new(big.Int)
	xb := new(big.Int)
	y := new(big.Int)

	x.SetBytes(document)
	n.Set(info.PublicKey.N)
	e.SetUint64(uint64(info.PublicKey.E))
	u.SetBytes(info.VerificationKey.U)

	// x = doc if (doc | n) == 1 else doc * u^e

	jacobied := false

	if big.Jacobi(x, n) == -1 {
		ue := new(big.Int).Exp(u, e, n)
		x.Mul(x, ue).Mod(x, n)
		jacobied = true
	}

	delta.MulRange(1, int64(info.L))
	ePrime.SetInt64(4).Mul(ePrime, delta).Mul(ePrime,delta)

	// Calculate w
	w.SetInt64(1)

	var i uint16
	for i = 0; i < k; i++ {
		id := int64(sigShares[i].Id)
		lambdaK2, err := sigShares.LagrangeInterpolation(id, int64(k), delta)
		if err != nil {
			return []byte{}, err
		}
		lambdaK2.Mul(lambdaK2, big.NewInt(2))
		si.SetBytes(sigShares[i].Xi)
		aux.Exp(si, lambdaK2, n)
		w.Mul(w, aux)
	}

	w.Mod(w, n)

	aux.GCD(a, b, ePrime, e)
	wa.Exp(w, a, n)
	xb.Exp(x, b, n)

	y.Mul(wa, xb)

	if jacobied {
		invU := new(big.Int).ModInverse(u, n)
		y.Mul(y, invU)
	}

	y.Mod(y, n)

	return y.Bytes(), nil

}

func (sigShares SignatureShares) LagrangeInterpolation(j, k int64, delta *big.Int) (*big.Int, error) {

	if int64(len(sigShares)) < k {
		return new(big.Int), fmt.Errorf("insufficient number of signature shares. provided: %d, needed: %d", len(sigShares), k)
	}

	out := new(big.Int).Set(delta)
	num := big.NewInt(1)
	den := big.NewInt(1)

	var i int64
	for i = 0; i < k; i++ {
		id := int64(sigShares[i].Id)
		if id != j {
			num.Mul(num, big.NewInt(id))   // num <-- num*j_
			den.Mul(den, big.NewInt(id-j)) // den <-- den*(j_-j)
		}
	}
	out.Mul(out, num)
	out.Quo(out, den)

	return out, nil
}
