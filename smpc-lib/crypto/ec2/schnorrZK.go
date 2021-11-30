/*
 *  Copyright (C) 2020-2021  AnySwap Ltd. All rights reserved.
 *  Copyright (C) 2020-2021  xing.chang@anyswap.exchange
 *
 *  This library is free software; you can redistribute it and/or
 *  modify it under the Apache License, Version 2.0.
 *
 *  This library is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
 *
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 *
 */

package ec2

import (
	"encoding/json"
	"fmt"
	"errors"
	"math/big"

	s256 "github.com/anyswap/Anyswap-MPCNode/crypto/secp256k1"
	"github.com/anyswap/Anyswap-MPCNode/internal/common/math/random"
)

//-------------------------------------------------------------------------------

// ZkUProof zku proof
type ZkUProof struct {
	E *big.Int
	S *big.Int
}

// ZkUProve create ZkUProof
func ZkUProve(u *big.Int) *ZkUProof {
	r := random.GetRandomIntFromZn(s256.S256().N)
	rGx, rGy := s256.S256().ScalarBaseMult(r.Bytes())
	uGx, uGy := s256.S256().ScalarBaseMult(u.Bytes())

	e := Sha512_256(rGx,rGy,uGx,uGy)

	s := new(big.Int).Mul(e, u)
	s = new(big.Int).Add(r, s)
	s = new(big.Int).Mod(s, s256.S256().N)

	zkUProof := &ZkUProof{E: e, S: s}
	return zkUProof
}

// ZkUVerify verify ZkUProof
func ZkUVerify(uG []*big.Int, zkUProof *ZkUProof) bool {
	// Check whether the point is on the curve
	if !checkPointOnCurve(uG) {
		return false
	}

	if zkUProof == nil || zkUProof.E == nil || zkUProof.S == nil {
	    return false
	}

	sGx, sGy := s256.S256().ScalarBaseMult(zkUProof.S.Bytes())

	minusE := new(big.Int).Mul(big.NewInt(-1), zkUProof.E)
	minusE = new(big.Int).Mod(minusE, s256.S256().N)

	eUx, eUy := s256.S256().ScalarMult(uG[0], uG[1], minusE.Bytes())
	rGx, rGy := s256.S256().Add(sGx, sGy, eUx, eUy)

	e := Sha512_256(rGx,rGy,uG[0],uG[1])
	if e.Cmp(zkUProof.E) == 0 {
		return true
	}
	
	return false
}

// MarshalJSON marshal ZkUProof to json bytes
func (zku *ZkUProof) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		E string `json:"E"`
		S string `json:"S"`
	}{
		E: fmt.Sprintf("%v", zku.E),
		S: fmt.Sprintf("%v", zku.S),
	})
}

// UnmarshalJSON unmarshal raw to ZkUProof
func (zku *ZkUProof) UnmarshalJSON(raw []byte) error {
	var zk struct {
		E string `json:"E"`
		S string `json:"S"`
	}
	if err := json.Unmarshal(raw, &zk); err != nil {
		return err
	}

	zku.E, _ = new(big.Int).SetString(zk.E, 10)
	zku.S, _ = new(big.Int).SetString(zk.S, 10)

	if zku.E == nil || zku.S == nil {
	    return errors.New("unmarshal json error")
	}

	return nil
}

//------------------------------------------------------------------------------------

// ZkXiProof the ZK that he knows xi using Schnorr’s protocol
type ZkXiProof struct {
	E *big.Int
	S *big.Int
}

// ZkXiProve create ZkXiProof
func ZkXiProve(sku1 *big.Int) *ZkXiProof {
	r := random.GetRandomIntFromZn(s256.S256().N)
	rGx, rGy := s256.S256().ScalarBaseMult(r.Bytes())
	xGx, xGy := s256.S256().ScalarBaseMult(sku1.Bytes())

	e := Sha512_256(rGx,rGy,xGx,xGy)

	s := new(big.Int).Mul(e, sku1)
	s = new(big.Int).Add(r, s)
	s = new(big.Int).Mod(s, s256.S256().N)

	zkxiProof := &ZkXiProof{E: e, S: s}
	return zkxiProof
}

// ZkXiVerify verify ZkXiProof
func ZkXiVerify(xiG []*big.Int, zkXiProof *ZkXiProof) bool {
	// Check whether the point is on the curve
	if !checkPointOnCurve(xiG) {
		return false
	}

	if zkXiProof.E == nil || zkXiProof.S == nil {
	    return false
	}

	sGx, sGy := s256.S256().ScalarBaseMult(zkXiProof.S.Bytes())

	minusE := new(big.Int).Mul(big.NewInt(-1), zkXiProof.E)
	minusE = new(big.Int).Mod(minusE, s256.S256().N)

	eUx, eUy := s256.S256().ScalarMult(xiG[0],xiG[1], minusE.Bytes())
	rGx, rGy := s256.S256().Add(sGx, sGy, eUx, eUy)

	e := Sha512_256(rGx,rGy,xiG[0],xiG[1])

	if e.Cmp(zkXiProof.E) == 0 {
		return true
	}
	
	return false
}

// MarshalJSON marshal ZkXiProof to json bytes
func (zkx *ZkXiProof) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		E string `json:"E"`
		S string `json:"S"`
	}{
		E: fmt.Sprintf("%v", zkx.E),
		S: fmt.Sprintf("%v", zkx.S),
	})
}

// UnmarshalJSON unmarshal raw to ZkXiProof
func (zkx *ZkXiProof) UnmarshalJSON(raw []byte) error {
	var zk struct {
		E string `json:"E"`
		S string `json:"S"`
	}
	if err := json.Unmarshal(raw, &zk); err != nil {
		return err
	}

	zkx.E, _ = new(big.Int).SetString(zk.E, 10)
	zkx.S, _ = new(big.Int).SetString(zk.S, 10)

	if zkx.E == nil || zkx.S == nil {
	    return errors.New("unmarshal json error")
	}

	return nil
}


