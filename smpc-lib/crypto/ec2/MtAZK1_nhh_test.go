package ec2_test

import (
    	"testing"

	"github.com/anyswap/Anyswap-MPCNode/crypto/secp256k1"
	"github.com/anyswap/Anyswap-MPCNode/internal/common/math/random"
	"github.com/anyswap/Anyswap-MPCNode/smpc-lib/crypto/ec2"
	"github.com/stretchr/testify/assert"
)

var (
    testNtildeLength = 2048
    testPaillierKeyLength = 2048
)

func TestMtAZK1Verify_nhh(t *testing.T) {
	publicKey,privateKey := ec2.CreatPair(testPaillierKeyLength)
	assert.NotZero(t, publicKey)
	assert.NotZero(t, privateKey)
	u1K := random.GetRandomIntFromZn(secp256k1.S256().N)
	nt,_,_,_,_ := ec2.CreateNt(testNtildeLength)
	assert.NotZero(t, nt)
	u1KCipher,u1R,_ := publicKey.Encrypt(u1K)
	u1u1MtAZK1Proof := ec2.MtAZK1Prove_nhh(u1K,u1R,publicKey,nt)
	assert.NotZero(t, u1u1MtAZK1Proof)
	u1rlt1 := u1u1MtAZK1Proof.MtAZK1Verify_nhh(u1KCipher,publicKey,nt)
	assert.True(t, u1rlt1, "u1rlt1 must be true")
}


