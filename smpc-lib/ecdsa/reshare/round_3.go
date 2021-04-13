package reshare 

import (
	"errors"
	"fmt"
	"math/big"
	"github.com/anyswap/Anyswap-MPCNode/smpc-lib/smpc"
	"github.com/anyswap/Anyswap-MPCNode/smpc-lib/crypto/ec2"
	"github.com/anyswap/Anyswap-MPCNode/crypto/secp256k1"
)

func (round *round3) Start() error {
	if round.started {
	    return errors.New("round already started")
	}
	round.number = 3
	round.started = true
	round.resetOK()

	//use round.temp.reshareRound1Messages replace round.idreshare,because round.idreshare == nil when oldnode == false
	for k,_ := range round.temp.reshareRound1Messages {
	    msg2,ok := round.temp.reshareRound2Messages[k].(*ReshareRound2Message)
	    if !ok {
		return errors.New("round.Start get round2 msg fail")
	    }
	   
	    ushare := &ec2.ShareStruct2{Id: msg2.Id, Share: msg2.Share}
	    msg21,ok := round.temp.reshareRound2Messages1[k].(*ReshareRound2Message1)
	    if !ok {
		return errors.New("round.Start get round2-1 msg fail")
	    }
	   
	    ps := &ec2.PolyGStruct2{PolyG: msg21.SkP1PolyG}
	    if !ushare.Verify2(ps) {
		fmt.Printf("========= round3 verify share fail, k = %v ==========\n",k)
		return errors.New("verify share data fail")
	    }
	    
	    //verify commitment
	    msg1,ok := round.temp.reshareRound1Messages[k].(*ReshareRound1Message)
	    if !ok {
		return errors.New("round.Start get round1 msg fail")
	    }
	    
	    deCommit := &ec2.Commitment{C: msg1.ComC, D: msg21.ComD}
	    if !deCommit.Verify() {
		fmt.Printf("========= round3 verify commitment fail, k = %v ==========\n",k)
		return errors.New("verify commitment fail")
	    }
	}
	
	var pkx *big.Int
	var pky *big.Int
	var newskU1 *big.Int

	for k,_ := range round.temp.reshareRound1Messages {
	    msg21,_ := round.temp.reshareRound2Messages1[k].(*ReshareRound2Message1)
	    msg1,_ := round.temp.reshareRound1Messages[k].(*ReshareRound1Message)
	    msg2,_ := round.temp.reshareRound2Messages[k].(*ReshareRound2Message)
	    ushare := &ec2.ShareStruct2{Id: msg2.Id, Share: msg2.Share}
	    
	    deCommit := &ec2.Commitment{C: msg1.ComC, D: msg21.ComD}
	    _, u1G := deCommit.DeCommit()
	    pkx = u1G[0]
	    pky = u1G[1]

	    newskU1 = ushare.Share
	    break
	}
	
	for k,_ := range round.temp.reshareRound1Messages {
	    if k == 0 {
		continue
	    }

	    msg2,_ := round.temp.reshareRound2Messages[k].(*ReshareRound2Message)
	    ushare := &ec2.ShareStruct2{Id: msg2.Id, Share: msg2.Share}
	    msg21,_ := round.temp.reshareRound2Messages1[k].(*ReshareRound2Message1)
	    msg1,_ := round.temp.reshareRound1Messages[k].(*ReshareRound1Message)
	    
	    deCommit := &ec2.Commitment{C: msg1.ComC, D: msg21.ComD}

	    _, u1G := deCommit.DeCommit()
	    pkx, pky = secp256k1.S256().Add(pkx, pky, u1G[0], u1G[1])

	    newskU1 = new(big.Int).Add(newskU1, ushare.Share)
	}
	
	newskU1 = new(big.Int).Mod(newskU1, secp256k1.S256().N)
	
	round.Save.SkU1 = newskU1
	round.Save.Pkx = pkx
	round.Save.Pky = pky

	idtmp,ok := new(big.Int).SetString(round.dnodeid,10)
	if !ok {
	    return errors.New("get id big number fail.")
	}

	cur_index := -1
	for k,v := range round.Save.Ids {
	    if v.Cmp(idtmp) == 0 {
		cur_index = k
		break
	    }
	}

	if cur_index < 0 {
	    return errors.New("get cur index fail")
	}

	u1PaillierPk, u1PaillierSk := ec2.GenerateKeyPair(round.paillierkeylength)

	round.Save.U1PaillierSk = u1PaillierSk
	round.Save.U1PaillierPk[cur_index] = u1PaillierPk
	
	re := &ReshareRound3Message{
	    ReshareRoundMessage:new(ReshareRoundMessage),
	    U1PaillierPk:u1PaillierPk,
	}
	re.SetFromID(round.dnodeid)
	re.SetFromIndex(cur_index)
	round.temp.reshareRound3Messages[cur_index] = re
	round.out <-re

	fmt.Printf("========= round3 start success ==========\n")
	return nil
}

func (round *round3) CanAccept(msg smpc.Message) bool {
	if _, ok := msg.(*ReshareRound3Message); ok {
		return msg.IsBroadcast()
	}
	return false
}

func (round *round3) Update() (bool, error) {
	for j, msg := range round.temp.reshareRound3Messages {
		if round.ok[j] {
			continue
		}
		if msg == nil || !round.CanAccept(msg) {
			return false, nil
		}
		
		round.ok[j] = true

		//add for reshare only
		if j == ( len(round.temp.reshareRound3Messages) - 1 ) {
		    for jj,_ := range round.ok {
			round.ok[jj] = true
		    }
		}
		//
	}
	
	return true, nil
}

func (round *round3) NextRound() smpc.Round {
	round.started = false
	return &round4{round}
}

