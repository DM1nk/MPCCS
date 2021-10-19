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
	"github.com/anyswap/Anyswap-MPCNode/smpc-lib/crypto/ec2"
	"github.com/anyswap/Anyswap-MPCNode/smpc-lib/smpc"
)

const (
	ntildeBitsLen = 2048
)

// Start broacast zku proof data
func (round *round5) Start() error {
	if round.started {
		return errors.New("round already started")
	}
	round.number = 5
	round.started = true
	round.resetOK()

	curIndex, err := round.GetDNodeIDIndex(round.dnodeid)
	if err != nil {
		return err
	}

	//check Ntilde bitlen
	for _,msg := range round.temp.kgRound4Messages {
		m,ok := msg.(*KGRound4Message)
		if !ok {
			return errors.New("error kg round4 message")
		}

		ntilde := m.U1NtildeH1H2
		if ntilde == nil || ntilde.Ntilde == nil {
			return errors.New("error kg round4 message")
		}

		if ntilde.Ntilde.BitLen() < ntildeBitsLen {
			return errors.New("got ntilde with not enough bits")
		}
	}
	//

	u1zkUProof := ec2.ZkUProve(round.temp.u1)
	if u1zkUProof == nil {
		return errors.New("zku prove fail")
	}

	kg := &KGRound5Message{
		KGRoundMessage: new(KGRoundMessage),
		U1zkUProof:     u1zkUProof,
	}
	kg.SetFromID(round.dnodeid)
	kg.SetFromIndex(curIndex)

	round.temp.kgRound5Messages[curIndex] = kg
	round.out <- kg

	//fmt.Printf("========= round5 start success ==========\n")
	return nil
}

// CanAccept is it legal to receive this message 
func (round *round5) CanAccept(msg smpc.Message) bool {
	if _, ok := msg.(*KGRound5Message); ok {
		return msg.IsBroadcast()
	}
	return false
}

// Update  is the message received and ready for the next round? 
func (round *round5) Update() (bool, error) {
	for j, msg := range round.temp.kgRound5Messages {
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
func (round *round5) NextRound() smpc.Round {
	round.started = false
	return &round6{round}
}
