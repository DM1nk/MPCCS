package smpc 

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"strings"
	"strconv"
	"encoding/json"
	"time"
	smpclib "github.com/anyswap/Anyswap-MPCNode/smpc-lib/smpc"
	"github.com/anyswap/Anyswap-MPCNode/smpc-lib/ecdsa/keygen"
	edkeygen "github.com/anyswap/Anyswap-MPCNode/smpc-lib/eddsa/keygen"
	"github.com/anyswap/Anyswap-MPCNode/internal/common"
	"github.com/anyswap/Anyswap-MPCNode/smpc-lib/crypto/ec2"
)

//---------------------------------------ECDSA start-----------------------------------------------------------------------

func ProcessInboundMessages(msgprex string,finishChan chan struct{},wg *sync.WaitGroup,ch chan interface{}) {
    defer wg.Done()
    fmt.Printf("start processing inbound messages\n")
    w, err := FindWorker(msgprex)
    if w == nil || err != nil {
	res := RpcSmpcRes{Ret: "", Err: fmt.Errorf("fail to process inbound messages")}
	ch <- res
	return 
    }

    defer fmt.Printf("stop processing inbound messages\n")
    for {
	    select {
	    case <-finishChan:
		    return
	    case m := <- w.SmpcMsg:

		    msgmap := make(map[string]string)
		    err := json.Unmarshal([]byte(m), &msgmap)
		    if err != nil {
			res := RpcSmpcRes{Ret:"",Err:err}
			ch <- res
			return
		    }
		    
		    if msgmap["Type"] == "KGRound0Message" { //0 message
			from := msgmap["FromID"]
			id,_ := new(big.Int).SetString(from,10)
			w.MsgToEnode[fmt.Sprintf("%v",id)] = msgmap["ENode"]
		    }

		    mm := GetRealMessage(msgmap) 
		    if mm == nil {
			res := RpcSmpcRes{Ret: "", Err: fmt.Errorf("fail to process inbound messages")}
			ch <- res
			return
		    }

		    _,err = w.DNode.Update(mm)
		    if err != nil {
			common.Error("====================ProcessInboundMessages,dnode update fail=======================","receiv msg",m,"err",err)
			res := RpcSmpcRes{Ret: "", Err: err}
			ch <- res
			return
		    }
	    }
    }
}

func GetRealMessage(msg map[string]string) smpclib.Message {
    from := msg["FromID"]
    var to []string
    v,ok := msg["ToID"]
    if ok && v != "" {
	to = strings.Split(v,":") 
    }

    index,indexerr := strconv.Atoi(msg["FromIndex"])
    if indexerr != nil {
	return nil
    }

    //1 message
    if msg["Type"] == "KGRound1Message" {
	pub := &ec2.PublicKey{}
	err := pub.UnmarshalJSON([]byte(msg["U1PaillierPk"]))
	if err == nil {
	    comc,_ := new(big.Int).SetString(msg["ComC"],10)
	    comc_bip32,_ := new(big.Int).SetString(msg["ComC_bip32"],10)
	    kg := &keygen.KGRound1Message{
		KGRoundMessage:new(keygen.KGRoundMessage),
		ComC:comc,
		ComC_bip32:comc_bip32,
		U1PaillierPk:pub,
	    }
	    kg.SetFromID(from)
	    kg.SetFromIndex(index)
	    kg.ToID = to
	    return kg
	}	
    }

    //2 message
    if msg["Type"] == "KGRound2Message" {
	id,_ := new(big.Int).SetString(msg["Id"],10)
	sh,_ := new(big.Int).SetString(msg["Share"],10)
	kg := &keygen.KGRound2Message{
	    KGRoundMessage:new(keygen.KGRoundMessage),
	    Id:id,
	    Share:sh,
	}
	kg.SetFromID(from)
	kg.SetFromIndex(index)
	kg.ToID = to
	return kg
    }

    //2-1 message
    if msg["Type"] == "KGRound2Message1" {
	c1,_ := new(big.Int).SetString(msg["C1"],10)
	kg := &keygen.KGRound2Message1{
	    KGRoundMessage:new(keygen.KGRoundMessage),
	    C1:c1,
	}
	kg.SetFromID(from)
	kg.SetFromIndex(index)
	kg.ToID = to
	return kg
    }

    //3 message
    if msg["Type"] == "KGRound3Message" {
	ugd := strings.Split(msg["ComU1GD"],":")
	u1gd := make([]*big.Int,len(ugd))
	for k,v := range ugd {
	    u1gd[k],_ = new(big.Int).SetString(v,10)
	}

	ucd := strings.Split(msg["ComC1GD"],":")
	u1cd := make([]*big.Int,len(ucd))
	for k,v := range ucd {
	    u1cd[k],_ = new(big.Int).SetString(v,10)
	}

	uggtmp := strings.Split(msg["U1PolyGG"],"|")
	ugg := make([][]*big.Int,len(uggtmp))
	for k,v := range uggtmp {
	    uggtmp2 := strings.Split(v,":")
	    tmp := make([]*big.Int,len(uggtmp2))
	    for kk,vv := range uggtmp2 {
		tmp[kk],_ = new(big.Int).SetString(vv,10)
	    }
	    ugg[k] = tmp
	}
	
	kg := &keygen.KGRound3Message{
	    KGRoundMessage:new(keygen.KGRoundMessage),
	    ComU1GD:u1gd,
	    ComC1GD:u1cd,
	    U1PolyGG:ugg,
	}
	kg.SetFromID(from)
	kg.SetFromIndex(index)
	kg.ToID = to
	return kg
    }

    //4 message
    if msg["Type"] == "KGRound4Message" {
	nti := &ec2.NtildeH1H2{}
	if err := nti.UnmarshalJSON([]byte(msg["U1NtildeH1H2"]));err == nil {
	    pf1 := &ec2.NtildeProof{}
	    if err := pf1.UnmarshalJSON([]byte(msg["NtildeProof1"]));err == nil {
		pf2 := &ec2.NtildeProof{}
		if err := pf2.UnmarshalJSON([]byte(msg["NtildeProof2"]));err == nil {
		    kg := &keygen.KGRound4Message{
			KGRoundMessage:new(keygen.KGRoundMessage),
			U1NtildeH1H2:nti,
			NtildeProof1:pf1,
			NtildeProof2:pf2,
		    }
		    kg.SetFromID(from)
		    kg.SetFromIndex(index)
		    kg.ToID = to
		    return kg
		}
	    }
	}
    }

    //5 message
    if msg["Type"] == "KGRound5Message" {
	zk := &ec2.ZkUProof{}
	if err := zk.UnmarshalJSON([]byte(msg["U1zkUProof"]));err == nil {
	    kg := &keygen.KGRound5Message{
		KGRoundMessage:new(keygen.KGRoundMessage),
		U1zkUProof:zk,
	    }
	    kg.SetFromID(from)
	    kg.SetFromIndex(index)
	    kg.ToID = to
	    return kg
	}
    }

    //6 message
    if msg["Type"] == "KGRound6Message" {
	b := false
	if msg["Check_Pubkey_Status"] == "true" {
	    b = true
	}

	kg := &keygen.KGRound6Message{
	    KGRoundMessage:new(keygen.KGRoundMessage),
	    Check_Pubkey_Status:b,
	}
	kg.SetFromID(from)
	kg.SetFromIndex(index)
	kg.ToID = to
	return kg
    }

    kg := &keygen.KGRound0Message{
	KGRoundMessage: new(keygen.KGRoundMessage),
    }
    kg.SetFromID(from)
    kg.SetFromIndex(-1)
    kg.ToID = to
    
    return kg 
}

func processKeyGen(msgprex string,errChan chan struct{},outCh <-chan smpclib.Message,endCh <-chan keygen.LocalDNodeSaveData) error {
	for {
		select {
		case <-errChan: // when keyGenParty return
		    fmt.Printf("=========== processKeyGen,error channel closed fail to start local smpc node, key = %v ===========\n",msgprex)
			return errors.New("error channel closed fail to start local smpc node")

		case <-time.After(time.Second * 300):
		    fmt.Printf("=========== processKeyGen,keygen timeout, key = %v ===========\n",msgprex)
			return errors.New("keygen timeout") 
		case msg := <-outCh:
			err := ProcessOutCh(msgprex,msg)
			if err != nil {
			    fmt.Printf("================ processKeyGen,process outch err = %v, key = %v ================\n",err,msgprex)
				return err
			}
		case msg := <-endCh:
			w, err := FindWorker(msgprex)
			if w == nil || err != nil {
			    return fmt.Errorf("get worker fail")
			}

			w.pkx.PushBack(fmt.Sprintf("%v",msg.Pkx))
			w.pky.PushBack(fmt.Sprintf("%v",msg.Pky))
			w.bip32c.PushBack(string(msg.C.Bytes()))
			w.sku1.PushBack(string(msg.SkU1.Bytes()))
			fmt.Printf("\n=================keygen finished successfully, pkx = %v,pky = %v,key = %v =================\n",msg.Pkx,msg.Pky,msgprex)

			ss := "XXX"
			ss = ss + common.SepSave
			s1 := msg.U1PaillierSk.Length
			s2 := string(msg.U1PaillierSk.L.Bytes())
			s3 := string(msg.U1PaillierSk.U.Bytes())
			ss = ss + s1 + common.SepSave + s2 + common.SepSave + s3 + common.SepSave
			
			for _,v := range msg.U1PaillierPk {
			    s1 = v.Length
			    s2 = string(v.N.Bytes())
			    s3 = string(v.G.Bytes())
			    s4 := string(v.N2.Bytes())
			    ss = ss + s1 + common.SepSave + s2 + common.SepSave + s3 + common.SepSave + s4 + common.SepSave
			}

			for _,v := range msg.U1NtildeH1H2 {
			    s1 = string(v.Ntilde.Bytes())
			    s2 = string(v.H1.Bytes())
			    s3 = string(v.H2.Bytes())
			    ss = ss + s1 + common.SepSave + s2 + common.SepSave + s3 + common.SepSave
			}

			ss += "NULL"
			w.save.PushBack(string(ss))

			return nil
		}
	}
}

type KGLocalDBSaveData struct {
    Save *keygen.LocalDNodeSaveData
    MsgToEnode map[string]string
}

func (kgsave *KGLocalDBSaveData) OutMap() map[string]string {
    out := kgsave.Save.OutMap()
    for key,value := range kgsave.MsgToEnode {
	out[key] = value
    }

    return out
}

func GetKGLocalDBSaveData(data map[string]string) *KGLocalDBSaveData {
    save := keygen.GetLocalDNodeSaveData(data)
    msgtoenode := make(map[string]string)
    for _,v := range save.Ids {
	msgtoenode[fmt.Sprintf("%v",v)] = data[fmt.Sprintf("%v",v)]
    }

    kgsave := &KGLocalDBSaveData{Save:save,MsgToEnode:msgtoenode}
    return kgsave
}

//---------------------------------------ECDSA end-----------------------------------------------------------------------


//---------------------------------------EDDSA start-----------------------------------------------------------------------

func ProcessInboundMessages_EDDSA(msgprex string,finishChan chan struct{},wg *sync.WaitGroup,ch chan interface{}) {
    defer wg.Done()
    fmt.Printf("start processing inbound messages, key = %v \n",msgprex)
    w, err := FindWorker(msgprex)
    if w == nil || err != nil {
	res := RpcSmpcRes{Ret: "", Err: fmt.Errorf("ed,fail to process inbound messages")}
	ch <- res
	return 
    }

    defer fmt.Printf("stop processing inbound messages, key = %v \n",msgprex)
    for {
	    select {
	    case <-finishChan:
		    return
	    case m := <- w.SmpcMsg:

		    msgmap := make(map[string]string)
		    err := json.Unmarshal([]byte(m), &msgmap)
		    if err != nil {
			res := RpcSmpcRes{Ret:"",Err:err}
			ch <- res
			return
		    }
		    
		    if msgmap["Type"] == "KGRound0Message" { //0 message
			from := msgmap["FromID"]
			w.MsgToEnode[from] = msgmap["ENode"]
		    }

		    mm := GetRealMessage_EDDSA(msgmap)
		    if mm == nil {
			res := RpcSmpcRes{Ret: "", Err: fmt.Errorf("ed,fail to process inbound messages")}
			ch <- res
			return
		    }

		    _,err = w.DNode.Update(mm)
		    if err != nil {
			common.Error("====================ProcessInboundMessages_EDDSA,dnode update fail=======================","receiv msg",m,"err",err)
			res := RpcSmpcRes{Ret: "", Err: err}
			ch <- res
			return
		    }
	    }
    }
}

func GetRealMessage_EDDSA(msg map[string]string) smpclib.Message {
    from := msg["FromID"]
    var to []string
    v,ok := msg["ToID"]
    if ok && v != "" {
	to = strings.Split(v,":") 
    }

    index,indexerr := strconv.Atoi(msg["FromIndex"])
    if indexerr != nil {
	return nil
    }

    //1 message
    if msg["Type"] == "KGRound1Message" {
	cpks, _ := hex.DecodeString(msg["CPk"])
	var temCpk [32]byte
	copy(temCpk[:], cpks[:])
	kg := &edkeygen.KGRound1Message{
	    KGRoundMessage:new(edkeygen.KGRoundMessage),
	    CPk:temCpk,
	}
	kg.SetFromID(from)
	kg.SetFromIndex(index)
	kg.ToID = to
	return kg
    }

    //2 message
    if msg["Type"] == "KGRound2Message" {
	zkpks, _ := hex.DecodeString(msg["ZkPk"])
	var temzkpk [64]byte
	copy(temzkpk[:], zkpks[:])
	kg := &edkeygen.KGRound2Message{
	    KGRoundMessage:new(edkeygen.KGRoundMessage),
	    ZkPk:temzkpk,
	}
	kg.SetFromID(from)
	kg.SetFromIndex(index)
	kg.ToID = to
	return kg
    }

    //3 message
    if msg["Type"] == "KGRound3Message" {
	dpks, _ := hex.DecodeString(msg["DPk"])
	var temdpk [64]byte
	copy(temdpk[:], dpks[:])
	kg := &edkeygen.KGRound3Message{
	    KGRoundMessage:new(edkeygen.KGRoundMessage),
	    DPk:temdpk,
	}
	kg.SetFromID(from)
	kg.SetFromIndex(index)
	kg.ToID = to
	return kg
    }

    //4 message
    if msg["Type"] == "KGRound4Message" {
	shares, _ := hex.DecodeString(msg["Share"])
	var temsh [32]byte
	copy(temsh[:], shares[:])
	kg := &edkeygen.KGRound4Message{
	    KGRoundMessage:new(edkeygen.KGRoundMessage),
	    Share:temsh,
	}
	kg.SetFromID(from)
	kg.SetFromIndex(index)
	kg.ToID = to
	return kg
    }

    //5 message
    if msg["Type"] == "KGRound5Message" {
	tmp := strings.Split(msg["CfsBBytes"],":")
	tmp2 := make([][32]byte,len(tmp))
	for k,v := range tmp {
	    vv, _ := hex.DecodeString(v)
	    var tem [32]byte
	    copy(tem[:],vv[:])
	    tmp2[k] = tem
	}

	kg := &edkeygen.KGRound5Message{
	    KGRoundMessage:new(edkeygen.KGRoundMessage),
	    CfsBBytes:tmp2,
	}
	kg.SetFromID(from)
	kg.SetFromIndex(index)
	kg.ToID = to
	return kg
    }

    kg := &edkeygen.KGRound0Message{
	KGRoundMessage: new(edkeygen.KGRoundMessage),
    }
    kg.SetFromID(from)
    kg.SetFromIndex(-1)
    kg.ToID = to
    
    return kg 
}

func processKeyGen_EDDSA(msgprex string,errChan chan struct{},outCh <-chan smpclib.Message,endCh <-chan edkeygen.LocalDNodeSaveData) error {
	for {
		select {
		case <-errChan: // when keyGenParty return
		    fmt.Printf("=========== processKeyGen_EDDSA,error channel closed fail to start local smpc node, key = %v ===========\n",msgprex)
			return errors.New("error channel closed fail to start local smpc node")

		case <-time.After(time.Second * 300):
		    fmt.Printf("====================== processKeyGen_EDDSA,ed keygen timeout, key = %v ====================\n",msgprex)
			return errors.New("ed keygen timeout") 
		case msg := <-outCh:
			err := ProcessOutCh(msgprex,msg)
			if err != nil {
			    fmt.Printf("================= processKeyGen_EDDSA,process outch err = %v,key = %v ==========\n",err,msgprex)
				return err
			}
		case msg := <-endCh:
			w, err := FindWorker(msgprex)
			if w == nil || err != nil {
			    return fmt.Errorf("get worker fail")
			}

			w.edsku1.PushBack(string(msg.Sk[:]))
			w.edpk.PushBack(string(msg.FinalPkBytes[:]))
			
			s := "XXX" + common.Sep11 + string(msg.Pk[:]) + common.Sep11 + string(msg.TSk[:]) + common.Sep11 + string(msg.FinalPkBytes[:])
			w.edsave.PushBack(string(s))
			fmt.Printf("=======================processKeyGen_EDDSA,success finish ed keygen, key = %v =======================\n",msgprex)
			return nil
		}
	}
}

type KGLocalDBSaveData_ed struct {
    Save *edkeygen.LocalDNodeSaveData
    MsgToEnode map[string]string
}

func (kgsave *KGLocalDBSaveData_ed) OutMap() map[string]string {
    out := kgsave.Save.OutMap()
    for key,value := range kgsave.MsgToEnode {
	out[key] = value
    }

    return out
}

func GetKGLocalDBSaveData_ed(data map[string]string) *KGLocalDBSaveData_ed {
    save := edkeygen.GetLocalDNodeSaveData(data)
    msgtoenode := make(map[string]string)
    for _,v := range save.Ids {
	var tmp [32]byte
	copy(tmp[:],v.Bytes())
	id := strings.ToLower(hex.EncodeToString(tmp[:]))
	msgtoenode[id] = data[id]
    }

    kgsave := &KGLocalDBSaveData_ed{Save:save,MsgToEnode:msgtoenode}
    return kgsave
}

//---------------------------------------EDDSA end-----------------------------------------------------------------------

func ProcessOutCh(msgprex string,msg smpclib.Message) error {
    if msg == nil {
	return fmt.Errorf("smpc info error")
    }

    w, err := FindWorker(msgprex)
    if w == nil || err != nil {
	return fmt.Errorf("get worker fail")
    }

    msgmap := msg.OutMap()
    msgmap["Key"] = msgprex
    msgmap["ENode"] = cur_enode
    s,err := json.Marshal(msgmap)
    if err != nil {
	fmt.Printf("====================ProcessOutCh, marshal err = %v, key = %v ========================\n",err,msgprex)
	return err
    }

    if msg.IsBroadcast() {
	SendMsgToSmpcGroup(string(s), w.groupid)
    } else {
	for _,v := range msg.GetToID() {
	    enode := w.MsgToEnode[v]
	    _, enodes := GetGroup(w.groupid)
	    nodes := strings.Split(enodes, common.Sep2)
	    for _, node := range nodes {
		node2 := ParseNode(node)
		if strings.EqualFold(enode,node2) {
		    SendMsgToPeer(node,string(s))
		    break
		}
	    }
	}
    }

    return nil
}

