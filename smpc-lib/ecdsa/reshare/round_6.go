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

package reshare

import (
	"errors"
	"github.com/anyswap/Anyswap-MPCNode/smpc-lib/smpc"
	"math/big"
)

// Start return save data
func (round *round6) Start() error {
	if round.started {
		return errors.New("round already started")
	}
	round.number = 6
	round.started = true
	round.resetOK()

	idtmp, ok := new(big.Int).SetString(round.dnodeid, 10)
	if !ok {
		return errors.New("get id big number fail.")
	}

	cur_index := -1
	for k, v := range round.Save.Ids {
		if v.Cmp(idtmp) == 0 {
			cur_index = k
			break
		}
	}

	if cur_index < 0 {
		return errors.New("get cur index fail")
	}

	round.Save.SkU1 = round.temp.newskU1
	round.Save.Pkx = round.temp.pkx
	round.Save.Pky = round.temp.pky
	round.Save.U1PaillierSk = round.temp.u1PaillierSk
	round.Save.U1PaillierPk[cur_index] = round.temp.u1PaillierPk
	round.Save.U1NtildeH1H2[cur_index] = round.temp.u1NtildeH1H2

	round.end <- *round.Save

	//fmt.Printf("========= round6 start success ==========\n")
	return nil
}

// CanAccept end reshare
func (round *round6) CanAccept(msg smpc.Message) bool {
	return false
}

// Update end reshare
func (round *round6) Update() (bool, error) {
	return false, nil
}

// NextRound end reshare
func (round *round6) NextRound() smpc.Round {
	return nil
}
