package main

import (
	"crypto"
	"crypto/sha256"
	"encoding/base64"
	"flag"
	"fmt"
	"log"
)

var kLong uint64
var lLong uint64
var s int
var m string

func init() {
	flag.Uint64Var(&kLong, "k", 3, "Threshold size")
	flag.Uint64Var(&lLong, "l", 5, "Number of shares")
	flag.IntVar(&s, "s", 1024, "key size in bits")
	flag.StringVar(&m, "m", "hello world", "message")
}

func main() {
	flag.Parse()

	// check that k and l are less than 2^16-1
	if kLong >= 1<<16 -1 || kLong <= 0 {
		panic("k should be between 1 and 65535")
	}

	if lLong >= 1<<16 -1 || lLong <= 0 {
		panic("l should be between 1 and 65535")
	}

	k := uint16(kLong)
	l := uint16(lLong)
	log.Printf("Generating %d keys for %d-threshold signing", l, k)
	keyShares, keyMeta, err := GenerateKeys(s, uint16(k), uint16(l), &KeyMetaArgs{})
	if err != nil {
		panic(fmt.Sprintf("%v", err))
	}
	log.Printf("Preparing document hash for string \"%s\"", m)
	docHash := sha256.Sum256([]byte(m))

	hashB64 := base64.StdEncoding.EncodeToString(docHash[:])
	log.Printf("The document hash is %s", hashB64)

	docPKCS1, err := PrepareDocumentHash(keyMeta.PublicKey.Size(), crypto.SHA256, docHash[:])
	if err != nil {
		panic(fmt.Sprintf("%v", err))
	}

	docB64 := base64.StdEncoding.EncodeToString(docPKCS1)

	log.Printf("The document pkcs1 prepared string is %s", docB64)

	sigs := make(SignatureShares, l)

	var i uint16
	for i = 0; i < l; i++ {
		log.Printf("signing with node %d of %d", keyShares[i].Id, l)
		sigs[i], err = keyShares[i].SignNode(docPKCS1, keyMeta)
		if err != nil {
			panic(fmt.Sprintf("%v", err))
		}
		log.Printf("verifying with node %d of %d", sigs[i].Id, l)
		if !sigs[i].Verify(docPKCS1, keyMeta) {
			panic("signature doesn't match")
		}
	}
	log.Printf("joining signatures")
	signature, err := sigs.Join(docPKCS1, keyMeta)
	if err != nil {
		panic(fmt.Sprintf("%v", err))
	}

	log.Printf("verifying signature")
	if !signature.Verify(docPKCS1, keyMeta) {
		panic(fmt.Sprintf("%v", err))
	}
	log.Printf("done!")
}