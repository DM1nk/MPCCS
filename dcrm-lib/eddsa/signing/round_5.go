package signing 

import (
	"errors"
	"fmt"
	//"math/big"
	"github.com/anyswap/Anyswap-MPCNode/dcrm-lib/dcrm"
)

func (round *round5) Start() error {
	if round.started {
	    fmt.Printf("============= ed sign,round5.start fail =======\n")
	    return errors.New("round already started")
	}

	round.number = 5
	round.started = true
	round.resetOK()

	cur_index,err := round.GetDNodeIDIndex(round.kgid)
	if err != nil {
	    return err
	}

	srm := &SignRound5Message{
	    SignRoundMessage: new(SignRoundMessage),
	    DSB:round.temp.DSB,
	}
	srm.SetFromID(round.kgid)
	srm.SetFromIndex(cur_index)

	round.temp.signRound5Messages[cur_index] = srm
	round.out <-srm
    
	fmt.Printf("============= ed sign,round5.start success, current node id = %v =============\n",round.kgid)

	return nil
}

func (round *round5) CanAccept(msg dcrm.Message) bool {
	if _, ok := msg.(*SignRound5Message); ok {
		return msg.IsBroadcast()
	}
	return false
}

func (round *round5) Update() (bool, error) {
	for j, msg := range round.temp.signRound5Messages {
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

func (round *round5) NextRound() dcrm.Round {
    //fmt.Printf("========= round.next round ========\n")
    round.started = false
    return &round6{round}
}

