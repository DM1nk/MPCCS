package keygen

import (
	"errors"
	"fmt"
	"math/big"
	"github.com/anyswap/Anyswap-MPCNode/dcrm-lib/dcrm"
)

var (
	zero = big.NewInt(0)
)

func newRound0(save *LocalDNodeSaveData, temp *localTempData,out chan<- dcrm.Message, end chan<- LocalDNodeSaveData,dnodeid string,dnodecount int,threshold int) dcrm.Round {
    return &round0{
		&base{save,temp,out,end,make([]bool,dnodecount),false,0,dnodeid,dnodecount,threshold}}
}

func (round *round0) Start() error {
	if round.started {
	    fmt.Printf("============= ed,round0.start fail =======\n")
	    return errors.New("round already started")
	}
	round.number = 0
	round.started = true
	round.resetOK()

	kg := &KGRound0Message{
	    KGRoundMessage: new(KGRoundMessage),
	}
	kg.SetFromID(round.dnodeid)
	kg.SetFromIndex(-1)

	round.temp.kgRound0Messages = append(round.temp.kgRound0Messages,kg)
	round.out <- kg
	fmt.Printf("============= round0.start success, current node id = %v =======\n",round.dnodeid)
	return nil
}

func (round *round0) CanAccept(msg dcrm.Message) bool {
	if _, ok := msg.(*KGRound0Message); ok {
		return msg.IsBroadcast()
	}
	return false
}

func (round *round0) Update() (bool, error) {
	for j, msg := range round.temp.kgRound0Messages {
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

func (round *round0) NextRound() dcrm.Round {
    fmt.Printf("========= round.next round ========\n")
    round.started = false
    return &round1{round}
}

