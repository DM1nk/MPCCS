package keygen

import (
	"errors"
	"github.com/anyswap/Anyswap-MPCNode/smpc-lib/smpc"
)

// Start broacast zkPk 
func (round *round2) Start() error {
	if round.started {
		return errors.New("ed,round already started")
	}
	round.number = 2
	round.started = true
	round.resetOK()

	ids, err := round.GetIds()
	if err != nil {
		return errors.New("ed,round.Start get ids fail.")
	}
	round.Save.Ids = ids

	cur_index, err := round.GetDNodeIDIndex(round.dnodeid)
	if err != nil {
		return err
	}
	round.Save.CurDNodeID = ids[cur_index]

	kg := &KGRound2Message{
		KGRoundMessage: new(KGRoundMessage),
		ZkPk:           round.temp.zkPk,
	}
	kg.SetFromID(round.dnodeid)
	kg.SetFromIndex(cur_index)
	round.temp.kgRound2Messages[cur_index] = kg
	round.out <- kg

	return nil
}

// CanAccept is it legal to receive this message 
func (round *round2) CanAccept(msg smpc.Message) bool {
	if _, ok := msg.(*KGRound2Message); ok {
		return msg.IsBroadcast()
	}
	return false
}

// Update  is the message received and ready for the next round? 
func (round *round2) Update() (bool, error) {
	for j, msg := range round.temp.kgRound2Messages {
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
func (round *round2) NextRound() smpc.Round {
	round.started = false
	return &round3{round}
}
