package ec2_test

import (
    	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/anyswap/Anyswap-MPCNode/smpc-lib/crypto/ec2"
	"math/big"
)

func TestCheckValidate(t *testing.T) {
    q := new(big.Int).SetInt64(5)
    p := ec2.GetP(q)
    sp := ec2.SafePrime{}
    sp.SetQ(q)
    sp.SetP(p)
    ret := sp.CheckValidate()
    assert.True(t,ret)
}

func TestCheckValidate_Bad(t *testing.T) {
    q := new(big.Int).SetInt64(10)
    p := ec2.GetP(q)
    sp := ec2.SafePrime{}
    sp.SetQ(q)
    sp.SetP(p)
    ret := sp.CheckValidate()
    assert.False(t,ret)
}



