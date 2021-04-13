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
	"github.com/anyswap/Anyswap-MPCNode/smpc-lib/ecdsa/reshare"
	"github.com/anyswap/Anyswap-MPCNode/internal/common"
	"github.com/anyswap/Anyswap-MPCNode/smpc-lib/crypto/ec2"
	"github.com/anyswap/Anyswap-MPCNode/crypto/secp256k1"
	"github.com/fsn-dev/cryptoCoins/coins"
	"github.com/anyswap/Anyswap-MPCNode/ethdb"
)

//ec

func ReshareProcessInboundMessages(msgprex string,finishChan chan struct{},wg *sync.WaitGroup,ch chan interface{}) {
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
		    
		    if msgmap["Type"] == "ReshareRound0Message" { //0 message
			from := msgmap["FromID"]
			id,_ := new(big.Int).SetString(from,10)
			w.MsgToEnode[fmt.Sprintf("%v",id)] = msgmap["ENode"]
		    }

		    mm := ReshareGetRealMessage(msgmap) 
		    if mm == nil {
			res := RpcSmpcRes{Ret: "", Err: fmt.Errorf("fail to process inbound messages")}
			ch <- res
			return
		    }

		    ld,ok := w.DNode.(*reshare.LocalDNode)
		    if ok && ld.CheckReshareMsg0(mm) {
			idreshare := GetIdReshareByGroupId(w.MsgToEnode,w.groupid)
			ld.SetIdReshare(idreshare)
			fmt.Printf("====================== ReshareProcessInboundMessages, check msg0, idreshare = %v, msgprex = %v ======================\n",idreshare,msgprex)
		    }

		    _,err = w.DNode.Update(mm)
		    if err != nil {
			fmt.Printf("========== ReshareProcessInboundMessages, dnode update fail, receiv smpc msg = %v, err = %v ============\n",m,err)
			res := RpcSmpcRes{Ret: "", Err: err}
			ch <- res
			return
		    }
	    }
    }
}

func ReshareGetRealMessage(msg map[string]string) smpclib.Message {
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
    if msg["Type"] == "ReshareRound1Message" {
	comc,_ := new(big.Int).SetString(msg["ComC"],10)
	re := &reshare.ReshareRound1Message{
	    ReshareRoundMessage:new(reshare.ReshareRoundMessage),
	    ComC:comc,
	}
	re.SetFromID(from)
	re.SetFromIndex(index)
	re.ToID = to
	return re
    }

    //2 message
    if msg["Type"] == "ReshareRound2Message" {
	id,_ := new(big.Int).SetString(msg["Id"],10)
	sh,_ := new(big.Int).SetString(msg["Share"],10)
	re := &reshare.ReshareRound2Message{
	    ReshareRoundMessage:new(reshare.ReshareRoundMessage),
	    Id:id,
	    Share:sh,
	}
	re.SetFromID(from)
	re.SetFromIndex(index)
	re.ToID = to
	fmt.Printf("============ GetRealMessage, get real message 2 success, share struct id = %v, share = %v, msg map = %v ===========\n",re.Id,re.Share,msg)
	return re 
    }

    //2-1 message
    if msg["Type"] == "ReshareRound2Message1" {
	ugd := strings.Split(msg["ComD"],":")
	u1gd := make([]*big.Int,len(ugd))
	for k,v := range ugd {
	    u1gd[k],_ = new(big.Int).SetString(v,10)
	}

	uggtmp := strings.Split(msg["SkP1PolyG"],"|")
	ugg := make([][]*big.Int,len(uggtmp))
	for k,v := range uggtmp {
	    uggtmp2 := strings.Split(v,":")
	    tmp := make([]*big.Int,len(uggtmp2))
	    for kk,vv := range uggtmp2 {
		tmp[kk],_ = new(big.Int).SetString(vv,10)
	    }
	    ugg[k] = tmp
	}
	
	re := &reshare.ReshareRound2Message1{
	    ReshareRoundMessage:new(reshare.ReshareRoundMessage),
	    ComD:u1gd,
	    SkP1PolyG:ugg,
	}
	re.SetFromID(from)
	re.SetFromIndex(index)
	re.ToID = to
	return re
    }

    //3 message
    if msg["Type"] == "ReshareRound3Message" {
	pub := &ec2.PublicKey{}
	err := pub.UnmarshalJSON([]byte(msg["U1PaillierPk"]))
	if err == nil {
	    fmt.Printf("============ ReshareGetRealMessage, get real message 3 success, msg map = %v ===========\n",msg)
	    re := &reshare.ReshareRound3Message{
		ReshareRoundMessage:new(reshare.ReshareRoundMessage),
		U1PaillierPk:pub,
	    }
	    re.SetFromID(from)
	    re.SetFromIndex(index)
	    re.ToID = to
	    return re 
	}
    }

    //4 message
    if msg["Type"] == "ReshareRound4Message" {
	nti := &ec2.NtildeH1H2{}
	if err := nti.UnmarshalJSON([]byte(msg["U1NtildeH1H2"]));err == nil {
	    fmt.Printf("============ ReshareGetRealMessage, get real message 4 success, msg map = %v ===========\n",msg)
	    re := &reshare.ReshareRound4Message{
		ReshareRoundMessage:new(reshare.ReshareRoundMessage),
		U1NtildeH1H2:nti,
	    }
	    re.SetFromID(from)
	    re.SetFromIndex(index)
	    re.ToID = to
	    return re
	}
    }

    fmt.Printf("============ ReshareGetRealMessage, get real message 0 success, msg map = %v ===========\n",msg)
    re := &reshare.ReshareRound0Message{
	ReshareRoundMessage: new(reshare.ReshareRoundMessage),
    }
    re.SetFromID(from)
    re.SetFromIndex(-1)
    re.ToID = to
    
    return re 
}

func processReshare(msgprex string,groupid string,pubkey string,account string,mode string,sigs string,errChan chan struct{},outCh <-chan smpclib.Message,endCh <-chan keygen.LocalDNodeSaveData) (*big.Int,error) {
	for {
		select {
		case <-errChan:
		fmt.Printf("=========== processReshare,error channel closed fail to start local smpc node ===========\n")
			return nil,errors.New("error channel closed fail to start local smpc node")

		case <-time.After(time.Second * 300):
		    fmt.Printf("=========== processReshare,reshare timeout ===========\n")
			// we bail out after KeyGenTimeoutSeconds
			return nil,errors.New("reshare timeout") 
		case msg := <-outCh:
			err := ReshareProcessOutCh(msgprex,groupid,msg)
			if err != nil {
			    fmt.Printf("======== processReshare,process outch err = %v ==========\n",err)
				return nil,err
			}
		case msg := <-endCh:
			w, err := FindWorker(msgprex)
			if w == nil || err != nil {
			    return nil,fmt.Errorf("get worker fail")
			}

			w.pkx.PushBack(fmt.Sprintf("%v",msg.Pkx))
			w.pky.PushBack(fmt.Sprintf("%v",msg.Pky))
			w.sku1.PushBack(fmt.Sprintf("%v",msg.SkU1))
			fmt.Printf("\n===========reshare finished successfully, pkx = %v,pky = %v ===========\n",msg.Pkx,msg.Pky)

			kgsave := &KGLocalDBSaveData{Save:(&msg),MsgToEnode:w.MsgToEnode}
			sdout := kgsave.OutMap()
			s,err := json.Marshal(sdout)
			if err != nil {
			    return nil,err
			}

			w.save.PushBack(string(s))

			smpcpks,err := hex.DecodeString(pubkey)
			if err != nil {
			    return nil,err
			}

			ys := secp256k1.S256().Marshal(msg.Pkx, msg.Pky)
			pubkeyhex := hex.EncodeToString(ys)
			if !strings.EqualFold(pubkey,pubkeyhex) {
			    common.Info("===================== reshare fail,new pubkey != old pubkey ====================","old pubkey",pubkey,"new pubkey",pubkeyhex,"key",msgprex)
			    return nil,errors.New("reshare fail,old pubkey != new pubkey") 
			}

			//set new sk
			dir := GetSkU1Dir()
			dbsktmp, err := ethdb.NewLDBDatabase(dir, cache, handles)
			//bug
			if err != nil {
				for i := 0; i < 100; i++ {
					dbsktmp, err = ethdb.NewLDBDatabase(dir, cache, handles)
					if err == nil {
						break
					}

					time.Sleep(time.Duration(1000000))
				}
			}
			if err != nil {
			    //dbsk = nil
			} else {
			    dbsk = dbsktmp
			}

			sk := KeyData{Key: smpcpks[:], Data: string(msg.SkU1.Bytes())}
			SkU1Chan <- sk

			for _, ct := range coins.Cointypes {
				if strings.EqualFold(ct, "ALL") {
					continue
				}

				h := coins.NewCryptocoinHandler(ct)
				if h == nil {
					continue
				}
				ctaddr, err := h.PublicKeyToAddress(pubkeyhex)
				if err != nil {
					continue
				}

				key := Keccak256Hash([]byte(strings.ToLower(ctaddr))).Hex()
				sk = KeyData{Key: []byte(key), Data: string(msg.SkU1.Bytes())}
				SkU1Chan <- sk
			}
			//

			dir = GetDbDir()
			dbtmp, err := ethdb.NewLDBDatabase(dir, cache, handles)
			//bug
			if err != nil {
				for i := 0; i < 100; i++ {
					dbtmp, err = ethdb.NewLDBDatabase(dir, cache, handles)
					if err == nil {
						break
					}

					time.Sleep(time.Duration(1000000))
				}
			}
			if err != nil {
			    //dbsk = nil
			} else {
			    db = dbtmp
			}

			nonce,_,err := GetReqAddrNonce(account) //reqaddr nonce
			if err != nil {
			    nonce = "0"
			}

			//**************TODO***************
			//default EC256K1 for reshare
			//ED25519 is not ready!!!

			rk := Keccak256Hash([]byte(strings.ToLower(account + ":" + "EC256K1" + ":" + groupid + ":" + nonce + ":" + w.limitnum + ":" + mode))).Hex() //reqaddr key
			//**********************************

			tt := fmt.Sprintf("%v",time.Now().UnixNano()/1e6)
			pubs := &PubKeyData{Key:rk,Account:account, Pub: string(smpcpks[:]), Save: string(s), Nonce: nonce, GroupId: groupid, LimitNum: w.limitnum, Mode: mode,KeyGenTime:tt,RefReShareKeys:msgprex}
			epubs, err := Encode2(pubs)
			if err != nil {
			    return nil,errors.New("encode PubKeyData fail in req ec2 pubkey")
			}

			ss1, err := Compress([]byte(epubs))
			if err != nil {
			    return nil,errors.New("compress PubKeyData fail in req ec2 pubkey")
			}

			exsit,pda := GetPubKeyDataFromLocalDb(string(smpcpks[:]))
			if exsit {
			    daa,ok := pda.(*PubKeyData)
			    if ok {
				//check mode
				if daa.Mode != mode {
				    return nil,errors.New("check mode fail")
				}
				//

				//check account
				if !strings.EqualFold(account, daa.Account) {
				    return nil,errors.New("check accout fail")
				}
				//

				go LdbPubKeyData.DeleteMap(daa.Key)
				kd := KeyData{Key: []byte(daa.Key), Data: "CLEAN"}
				PubKeyDataChan <- kd
			    }
			}
			
			kd := KeyData{Key: smpcpks[:], Data: ss1}
			PubKeyDataChan <- kd
			/////
			LdbPubKeyData.WriteMap(string(smpcpks[:]), pubs)
			////
			for _, ct := range coins.Cointypes {
				if strings.EqualFold(ct, "ALL") {
					continue
				}

				h := coins.NewCryptocoinHandler(ct)
				if h == nil {
					continue
				}
				ctaddr, err := h.PublicKeyToAddress(pubkey)
				if err != nil {
					continue
				}

				key := Keccak256Hash([]byte(strings.ToLower(ctaddr))).Hex()
				kd = KeyData{Key: []byte(key), Data: ss1}
				PubKeyDataChan <- kd
				/////
				LdbPubKeyData.WriteMap(key, pubs)
				////
			}
			
			_,err = SetReqAddrNonce(account,nonce)
			if err != nil {
			    return nil,errors.New("set reqaddr nonce fail")
			}

			wid := -1
			var allreply []NodeReply
			exsit,da2 := GetValueFromPubKeyData(msgprex)
			if exsit {
			    acr,ok := da2.(*AcceptReShareData)
			    if ok {
				wid = acr.WorkId
				allreply = acr.AllReply
			    }
			}

			ac := &AcceptReqAddrData{Initiator:cur_enode,Account: account, Cointype: "EC256K1", GroupId: groupid, Nonce: nonce, LimitNum: w.limitnum, Mode: mode, TimeStamp: tt, Deal: "true", Accept: "true", Status: "Success", PubKey: pubkey, Tip: "", Error: "", AllReply: allreply, WorkId: wid,Sigs:sigs}
			err = SaveAcceptReqAddrData(ac)
			if err != nil {
			    return nil,errors.New("save reqaddr accept data fail")
			}

			if mode == "0" {
			    sigs2 := strings.Split(ac.Sigs,common.Sep)
			    cnt,_ := strconv.Atoi(sigs2[0])
			    for j := 0;j<cnt;j++ {
				fr := sigs2[2*j+2]
				exsit,da := GetValueFromPubKeyData(strings.ToLower(fr))
				if !exsit {
				    kdtmp := KeyData{Key: []byte(strings.ToLower(fr)), Data: rk}
				    PubKeyDataChan <- kdtmp
				    LdbPubKeyData.WriteMap(strings.ToLower(fr), []byte(rk))
				} else {
				    //
				    found := false
				    keys := strings.Split(string(da.([]byte)),":")
				    for _,v := range keys {
					if strings.EqualFold(v,rk) {
					    found = true
					    break
					}
				    }
				    //

				    if !found {
					da2 := string(da.([]byte)) + ":" + rk
					kdtmp := KeyData{Key: []byte(strings.ToLower(fr)), Data: da2}
					PubKeyDataChan <- kdtmp
					LdbPubKeyData.WriteMap(strings.ToLower(fr), []byte(da2))
				    }
				}
			    }
			} else {
			    exsit,da := GetValueFromPubKeyData(strings.ToLower(account))
			    if !exsit {
				kdtmp := KeyData{Key: []byte(strings.ToLower(account)), Data: rk}
				PubKeyDataChan <- kdtmp
				LdbPubKeyData.WriteMap(strings.ToLower(account), []byte(rk))
			    } else {
				//
				found := false
				keys := strings.Split(string(da.([]byte)),":")
				for _,v := range keys {
				    if strings.EqualFold(v,rk) {
					found = true
					break
				    }
				}
				//

				if !found {
				    da2 := string(da.([]byte)) + ":" + rk
				    kdtmp := KeyData{Key: []byte(strings.ToLower(account)), Data: da2}
				    PubKeyDataChan <- kdtmp
				    LdbPubKeyData.WriteMap(strings.ToLower(account), []byte(da2))
				}

				}
			}
			
			return msg.SkU1,nil
		}
	}
}

func ReshareProcessOutCh(msgprex string,groupid string,msg smpclib.Message) error {
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
	fmt.Printf("====================ReshareProcessOutCh, marshal err = %v ========================\n",err)
	return err
    }

    if msg.IsBroadcast() {
	fmt.Printf("=========== ReshareProcessOutCh,broacast msg = %v, group id = %v ===========\n",string(s),groupid)

	SendMsgToSmpcGroup(string(s), groupid)
    } else {
	for _,v := range msg.GetToID() {
	    fmt.Printf("===============ReshareProcessOutCh, to id = %v,groupid = %v ==============\n",v,groupid)
	    enode := w.MsgToEnode[v]
	    _, enodes := GetGroup(groupid)
	    nodes := strings.Split(enodes, common.Sep2)
	    for _, node := range nodes {
		node2 := ParseNode(node)
		//fmt.Printf("===============ReshareProcessOutCh, enode = %v,node2 = %v ==============\n",enode,node2)
		
		if strings.EqualFold(enode,node2) {
		    fmt.Printf("=========== ReshareProcessOutCh,send msg = %v, group id = %v,send to peer = %v ===========\n",string(s),groupid,node)
		    SendMsgToPeer(node,string(s))
		    break
		}
	    }
	}
    }

    return nil
}

