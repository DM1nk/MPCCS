/*
 *  Copyright (C) 2020-2021  AnySwap Ltd. All rights reserved.
 *  Copyright (C) 2020-2021  haijun.cai@anyswap.exchange
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

package keygen

import (
	"errors"
	"fmt"
	"github.com/anyswap/FastMulThreshold-DSA/smpc-lib/crypto/ed"
	"github.com/anyswap/FastMulThreshold-DSA/smpc-lib/smpc"
	cryptorand "crypto/rand"
	"io"
)

// Start get sk pk
func (round *round1) Start() error {
	if round.started {
		fmt.Printf("============ round1 start error,already started============\n")
		return errors.New("round already started")
	}
	round.number = 1
	round.started = true
	round.resetOK()

	//1.1-1.2 generate 32-bits privatekey', then bit calculation to privatekey
	rand := cryptorand.Reader

	var sk [32]byte
	var pk [32]byte
	var skTem [64]byte

	if _, err := io.ReadFull(rand, sk[:]); err != nil {
	    fmt.Println("Error: io.ReadFull(rand, sk)")
	}

	sk[0] &= 248
	sk[31] &= 127
	sk[31] |= 64

	copy(skTem[:], sk[:])
	ed.ScReduce(&sk, &skTem)
	
	// 1.3 publicKey

	var A ed.ExtendedGroupElement
	ed.GeScalarMultBase(&A, &sk)

	A.ToBytes(&pk)
	///////////////////solana
	/*var sk [64]byte
	var pk [32]byte

	seedDigest := sha512.Sum512(seed[:])

	seedDigest[0] &= 248
	seedDigest[31] &= 127
	seedDigest[31] |= 64

	var A ed.ExtendedGroupElement
	var temSk [32]byte
	copy(temSk[:], seedDigest[:])
	ed.GeScalarMultBase(&A, &temSk)
	A.ToBytes(&pk)

	copy(sk[:], seed[:])
	copy(sk[32:], pk[:])*/
	/////////////////solana

	CPk, DPk, err := ed.Commit(pk)
	if err != nil {
	    return err
	}

	zkPk,err := ed.Prove2(sk,pk)
	if err != nil {
	    return err
	}

	round.temp.sk = sk
	round.temp.pk = pk
	round.temp.DPk = DPk
	round.temp.zkPk = zkPk

	index, err := round.GetDNodeIDIndex(round.dnodeid)
	if err != nil {
		fmt.Printf("============round1 start,get dnode id index fail,uid = %v,err = %v ===========\n", round.dnodeid,err)
		return err
	}

	kg := &KGRound1Message{
		KGRoundMessage: new(KGRoundMessage),
		CPk:            CPk,
	}
	kg.SetFromID(round.dnodeid)
	kg.SetFromIndex(index)

	round.Save.Sk = sk
	round.Save.Pk = pk
	round.temp.kgRound1Messages[index] = kg
	round.out <- kg

	//fmt.Printf("============ round1 start success,index = %v ============\n",index)
	return nil
}

// CanAccept is it legal to receive this message 
func (round *round1) CanAccept(msg smpc.Message) bool {
	if _, ok := msg.(*KGRound1Message); ok {
		return msg.IsBroadcast()
	}
	return false
}

// Update  is the message received and ready for the next round? 
func (round *round1) Update() (bool, error) {
	for j, msg := range round.temp.kgRound1Messages {
		if round.ok[j] {
			continue
		}
		if msg == nil || !round.CanAccept(msg) {
			return false, nil
		}
		round.ok[j] = true
	}

	return true, nil
}

// NextRound enter next round
func (round *round1) NextRound() smpc.Round {
	round.started = false
	return &round2{round}
}
