package main

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

func CreatePrivatePublicKey() {
	privKey := secp256k1.GenPrivKey()
	privateKeyString := hex.EncodeToString(privKey.Bytes())
	finalPrivKey := base64.StdEncoding.EncodeToString([]byte(privateKeyString))
	fmt.Println(finalPrivKey)
}
