/*
 *  Copyright (C) 2018-2019  Fusion Foundation Ltd. All rights reserved.
 *  Copyright (C) 2018-2019  caihaijun@fusion.org
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

package smpc 

import (
	"github.com/anyswap/Anyswap-MPCNode/internal/common"
	"github.com/anyswap/Anyswap-MPCNode/crypto/secp256k1"
	"strings"
	"math/big"
	"encoding/hex"
	"fmt"
	"time"
	"container/list"
	"github.com/fsn-dev/cryptoCoins/coins"
	"crypto/ecdsa"
	"github.com/anyswap/Anyswap-MPCNode/crypto"
	"github.com/anyswap/Anyswap-MPCNode/crypto/ecies"
	"strconv"
	"github.com/anyswap/Anyswap-MPCNode/p2p/discover"
	crand "crypto/rand"
	"github.com/fsn-dev/cryptoCoins/coins/types"
	"github.com/fsn-dev/cryptoCoins/tools/rlp"
	"encoding/json"
	"crypto/hmac"
	"crypto/sha512"
)

var (
	ch_t                     = 300 
	WaitMsgTimeGG20                     = 100
	waitall                     = ch_t * recalc_times
	waitallgg20                     = WaitMsgTimeGG20 * recalc_times
	WaitAgree = 120
	C1Data = common.NewSafeMap(10) 
	
	//callback
	GetGroup               func(string) (int, string)
	SendToGroupAllNodes    func(string, string) (string, error)
	GetSelfEnode           func() string
	BroadcastInGroupOthers func(string, string) (string, error)
	SendToPeer             func(string, string) error
	ParseNode              func(string) string
	GetEosAccount          func() (string, string, string)
)

//p2p callback
func RegP2pGetGroupCallBack(f func(string) (int, string)) {
	GetGroup = f
}

func RegP2pSendToGroupAllNodesCallBack(f func(string, string) (string, error)) {
	SendToGroupAllNodes = f
}

func RegP2pGetSelfEnodeCallBack(f func() string) {
	GetSelfEnode = f
}

func RegP2pBroadcastInGroupOthersCallBack(f func(string, string) (string, error)) {
	BroadcastInGroupOthers = f
}

func RegP2pSendMsgToPeerCallBack(f func(string, string) error) {
	SendToPeer = f
}

func RegP2pParseNodeCallBack(f func(string) string) {
	ParseNode = f
}

func RegSmpcGetEosAccountCallBack(f func() (string, string, string)) {
	GetEosAccount = f
}

////////////////////////////////

func SendMsgToSmpcGroup(msg string, groupid string) {
	common.Debug("=========SendMsgToSmpcGroup=============","msg",msg,"groupid",groupid)
	_,err := BroadcastInGroupOthers(groupid, msg)
	if err != nil {
	    common.Debug("=========SendMsgToSmpcGroup,send msg to smpc group=============","msg",msg,"groupid",groupid,"err",err)
	}
}

///
func EncryptMsg(msg string, enodeID string) (string, error) {
	//fmt.Println("=============EncryptMsg,KeyFile = %s,enodeID = %s ================",KeyFile,enodeID)
	hprv, err1 := hex.DecodeString(enodeID)
	if err1 != nil {
		return "", err1
	}

	//fmt.Println("=============EncryptMsg,hprv len = %v ================",len(hprv))
	p := &ecdsa.PublicKey{Curve: crypto.S256(), X: new(big.Int), Y: new(big.Int)}
	half := len(hprv) / 2
	p.X.SetBytes(hprv[:half])
	p.Y.SetBytes(hprv[half:])
	if !p.Curve.IsOnCurve(p.X, p.Y) {
		return "", fmt.Errorf("id is invalid secp256k1 curve point")
	}

	var cm []byte
	pub := ecies.ImportECDSAPublic(p)
	cm, err := ecies.Encrypt(crand.Reader, pub, []byte(msg), nil, nil)
	if err != nil {
		return "", err
	}

	return string(cm), nil
}

func DecryptMsg(cm string) (string, error) {
	//test := Keccak256Hash([]byte(strings.ToLower(cm))).Hex()
	nodeKey, errkey := crypto.LoadECDSA(KeyFile)
	if errkey != nil {
		//fmt.Printf("%v =========DecryptMsg finish crypto.LoadECDSA,err = %v,keyfile = %v,msg hash = %v =================\n", common.CurrentTime(), errkey, KeyFile, test)
		return "", errkey
	}

	prv := ecies.ImportECDSA(nodeKey)
	var m []byte
	m, err := prv.Decrypt([]byte(cm), nil, nil)
	if err != nil {
		//fmt.Printf("%v =========DecryptMsg finish prv.Decrypt,err = %v,keyfile = %v,msg hash = %v =================\n", common.CurrentTime(), err, KeyFile, test)
		return "", err
	}

	return string(m), nil
}

///
func SendMsgToPeer(enodes string, msg string) {
//	common.Debug("=========SendMsgToPeer===========","msg",msg,"send to peer",enodes)
	en := strings.Split(string(enodes[8:]), "@")
	cm, err := EncryptMsg(msg, en[0])
	if err != nil {
		//fmt.Printf("%v =========SendMsgToPeer,encrypt msg fail,err = %v =================\n", common.CurrentTime(), err)
		return
	}

	err = SendToPeer(enodes, cm)
	if err != nil {
//	    common.Debug("=========SendMsgToPeer,send to peer fail===========","msg",msg,"send to peer",enodes,"err",err)
	    return
	}
}

type RawReply struct {
    From string
    Accept string
    TimeStamp string
}

func GetRawReply(l *list.List) *common.SafeMap {
    ret := common.NewSafeMap(10) 
    if l == nil {
	return ret
    }

    var next *list.Element
    for e := l.Front(); e != nil; e = next {
	next = e.Next()

	if e.Value == nil {
		continue
	}

	s := e.Value.(string)

	if s == "" {
		continue
	}

	raw := s 
	keytmp,from,_,txdata,err := CheckRaw(raw)
	if err != nil {
	    continue
	}
	
	req,ok := txdata.(*TxDataReqAddr)
	if ok {
	    reply := &RawReply{From:from,Accept:"true",TimeStamp:req.TimeStamp}
	    tmp,ok := ret.ReadMap(from)
	    if !ok {
		ret.WriteMap(from,reply)
	    } else {
		tmp2,ok := tmp.(*RawReply)
		if ok {
		    t1,_ := new(big.Int).SetString(reply.TimeStamp,10)
		    t2,_ := new(big.Int).SetString(tmp2.TimeStamp,10)
		    if t1.Cmp(t2) > 0 {
			ret.WriteMap(from,reply)
		    }
		}

	    }

	    continue
	}
	
	sig,ok := txdata.(*TxDataSign)
	if ok {
	    common.Debug("=================GetRawReply,the list item is TxDataSign=================","key",keytmp,"from",from,"sig",sig)
	    reply := &RawReply{From:from,Accept:"true",TimeStamp:sig.TimeStamp}
	    tmp,ok := ret.ReadMap(from)
	    if !ok {
		ret.WriteMap(from,reply)
	    } else {
		tmp2,ok := tmp.(*RawReply)
		if ok {
		    t1,_ := new(big.Int).SetString(reply.TimeStamp,10)
		    t2,_ := new(big.Int).SetString(tmp2.TimeStamp,10)
		    if t1.Cmp(t2) > 0 {
			ret.WriteMap(from,reply)
		    }
		}
	    }

	    continue
	}
	
	rh,ok := txdata.(*TxDataReShare)
	if ok {
	    reply := &RawReply{From:from,Accept:"true",TimeStamp:rh.TimeStamp}
	    tmp,ok := ret.ReadMap(from)
	    if !ok {
		ret.WriteMap(from,reply)
	    } else {
		tmp2,ok := tmp.(*RawReply)
		if ok {
		    t1,_ := new(big.Int).SetString(reply.TimeStamp,10)
		    t2,_ := new(big.Int).SetString(tmp2.TimeStamp,10)
		    if t1.Cmp(t2) > 0 {
			ret.WriteMap(from,reply)
		    }
		}
	    }

	    continue
	}
	
	acceptreq,ok := txdata.(*TxDataAcceptReqAddr)
	if ok {
	    accept := "false"
	    if acceptreq.Accept == "AGREE" {
		    accept = "true"
	    }

	    reply := &RawReply{From:from,Accept:accept,TimeStamp:acceptreq.TimeStamp}
	    tmp,ok := ret.ReadMap(from)
	    if !ok {
		ret.WriteMap(from,reply)
	    } else {
		tmp2,ok := tmp.(*RawReply)
		if ok {
		    t1,_ := new(big.Int).SetString(reply.TimeStamp,10)
		    t2,_ := new(big.Int).SetString(tmp2.TimeStamp,10)
		    if t1.Cmp(t2) > 0 {
			ret.WriteMap(from,reply)
		    }
		}

	    }
	}
	
	acceptsig,ok := txdata.(*TxDataAcceptSign)
	if ok {
	    common.Info("=================GetRawReply,the list item is TxDataAcceptSign================","key",keytmp,"from",from,"accept",acceptsig.Accept,"raw",raw)
	    accept := "false"
	    if acceptsig.Accept == "AGREE" {
		    accept = "true"
	    }

	    reply := &RawReply{From:from,Accept:accept,TimeStamp:acceptsig.TimeStamp}
	    tmp,ok := ret.ReadMap(from)
	    if !ok {
		ret.WriteMap(from,reply)
	    } else {
		tmp2,ok := tmp.(*RawReply)
		if ok {
		    t1,_ := new(big.Int).SetString(reply.TimeStamp,10)
		    t2,_ := new(big.Int).SetString(tmp2.TimeStamp,10)
		    if t1.Cmp(t2) > 0 {
			ret.WriteMap(from,reply)
		    }
		}

	    }
	}
	
	acceptrh,ok := txdata.(*TxDataAcceptReShare)
	if ok {
	    accept := "false"
	    if acceptrh.Accept == "AGREE" {
		    accept = "true"
	    }

	    reply := &RawReply{From:from,Accept:accept,TimeStamp:acceptrh.TimeStamp}
	    tmp,ok := ret.ReadMap(from)
	    if !ok {
		ret.WriteMap(from,reply)
	    } else {
		tmp2,ok := tmp.(*RawReply)
		if ok {
		    t1,_ := new(big.Int).SetString(reply.TimeStamp,10)
		    t2,_ := new(big.Int).SetString(tmp2.TimeStamp,10)
		    if t1.Cmp(t2) > 0 {
			ret.WriteMap(from,reply)
		    }
		}

	    }
	}
    }

    return ret
}

func CheckReply(l *list.List,rt RpcType,key string) bool {
    if l == nil || key == "" {
	return false
    }

    /////reshare only
    if rt == Rpc_RESHARE {
	exsit,da := GetReShareInfoData([]byte(key))
	if !exsit {
	    return false
	}

	ac,ok := da.(*AcceptReShareData)
	if !ok || ac == nil {
	    return false
	}

	ret := GetRawReply(l)
	_, enodes := GetGroup(ac.GroupId)
	nodes := strings.Split(enodes, common.Sep2)
	for _, node := range nodes {
	    node2 := ParseNode(node)
	    pk := "04" + node2 
	     h := coins.NewCryptocoinHandler("FSN")
	     if h == nil {
		continue
	     }

	    fr, err := h.PublicKeyToAddress(pk)
	    if err != nil {
		return false
	    }

	    found := false
	    _,value := ret.ListMap()
	    for _,v := range value {
		if v != nil && strings.EqualFold((v.(*RawReply)).From,fr) {
		    found = true
		    break
		}
	    }

	    if !found {
		return false
	    }
	}

	return true
    }
    /////////////////

    k := ""
    if rt == Rpc_REQADDR {
	k = key
    } else {
	k = GetReqAddrKeyByOtherKey(key,rt)
    }

    if k == "" {
	return false
    }

    exsit,da := GetReqAddrInfoData([]byte(k))
    if !exsit || da == nil {
	exsit,da = GetPubKeyData([]byte(k))
    }

    if !exsit {
	return false
    }

    ac,ok := da.(*AcceptReqAddrData)
    if !ok {
	return false
    }

    if ac == nil {
	return false
    }

    ret := GetRawReply(l)

    if rt == Rpc_REQADDR {
	//sigs:  5:eid1:acc1:eid2:acc2:eid3:acc3:eid4:acc4:eid5:acc5
	mms := strings.Split(ac.Sigs, common.Sep)
	count := (len(mms) - 1)/2
	if count <= 0 {
	    common.Debug("===================== CheckReply,reqaddr================","ac.Sigs",ac.Sigs,"count",count,"k",k,"key",key,"ret",ret)
	    return false
	}

	for j:=0;j<count;j++ {
	    found := false
	    _,value := ret.ListMap()
	    for _,v := range value {
		if v != nil && strings.EqualFold((v.(*RawReply)).From,mms[2*j+2]) { //allow user login diffrent node
		    found = true
		    break
		}
	    }

	    if !found {
		common.Debug("===================== CheckReply,reqaddr, return false.====================","ac.Sigs",ac.Sigs,"count",count,"k",k,"key",key)
		return false
	    }
	}

	return true
    }

    if rt == Rpc_SIGN {
	common.Debug("===================== CheckReply,get raw reply finish================","key",key)
	exsit,data := GetSignInfoData([]byte(key))
	if !exsit {
	    common.Debug("===================== CheckReply,get raw reply finish and get value by key fail================","key",key)
	    return false
	}

	sig,ok := data.(*AcceptSignData)
	if !ok || sig == nil {
    common.Debug("===================== CheckReply,get raw reply finish and get accept sign data by key fail================","key",key)
	    return false
	}

	mms := strings.Split(ac.Sigs, common.Sep)
	_, enodes := GetGroup(sig.GroupId)
	nodes := strings.Split(enodes, common.Sep2)
	for _, node := range nodes {
	    node2 := ParseNode(node)
	    foundeid := false
	    for kk,v := range mms {
		if strings.EqualFold(v,node2) {
		    foundeid = true
		    found := false
		    _,value := ret.ListMap()
		    for _,vv := range value {
			if vv != nil && strings.EqualFold((vv.(*RawReply)).From,mms[kk+1]) { //allow user login diffrent node
			    found = true
			    break
			}
		    }

		    if !found {
			common.Debug("===================== CheckReply,mms[kk+1] no find in ret map and return fail==================","key",key,"mms[kk+1]",mms[kk+1])
			return false
		    }

		    break
		}
	    }

	    if !foundeid {
	    common.Debug("===================== CheckReply,get raw reply finish and find eid fail================","key",key)
		return false
	    }
	}

	return true
    }

    return false 
}

//-------------------------------------------------------------

func GetCmdKey(msg string) string {
    if msg == "" {
	return ""
    }

    ok,key := IsGenKeyCmd(msg)
    if ok {
	return key
    }

    ok,key = IsReshareCmd(msg)
    if ok {
	return key
    }

    key,ok = IsPreGenSignData(msg)
    if ok {
	return key
    }

    key,ok = IsEDSignCmd(msg)
    if ok {
	return key
    }

    key,ok = IsSignDataCmd(msg)
    if ok {
	return key
    }

    return ""
}

func Call(msg interface{}, enode string) {
	common.Debug("====================Call===================","get p2p msg ",msg,"sender node",enode)
	s := msg.(string)
	if s == "" {
	    return
	}
	raw,err := UnCompress(s)
	if err == nil {
	    s = raw
	}
	msgdata, errdec := DecryptMsg(s) //for SendMsgToPeer
	if errdec == nil {
		s = msgdata
	}
	
	msgmap := make(map[string]string)
	err = json.Unmarshal([]byte(s), &msgmap)
	if err == nil {
	    val,ok := msgmap["Key"]
	    if ok {
		w, err := FindWorker(val)
		if err == nil {
		    if w.DNode != nil && w.DNode.Round() != nil {
			w.SmpcMsg <- s
		    } else {
			from := msgmap["FromID"]
			msgtype := msgmap["Type"]
			key := strings.ToLower(val + "-" + from + "-" + msgtype)
			C1Data.WriteMap(key,s)
			fmt.Printf("===============================Call, pre-save p2p msg, worker found, key = %v,fromId = %v,msgtype = %v, msg = %v========================\n",val,from,msgtype,s)
		    }
		} else {
		    from := msgmap["FromID"]
		    msgtype := msgmap["Type"]
		    key := strings.ToLower(val + "-" + from + "-" + msgtype)
		    C1Data.WriteMap(key,s)
		    fmt.Printf("===============================Call, pre-save p2p msg, worker not found, key = %v,fromId = %v,msgtype = %v, msg = %v========================\n",val,from,msgtype,s)
		}
		
		return
	    }
	}

	SetUpMsgList(s,enode)
}

func MergeAllPreSaveMsgToWorkerId(wid int) {
    if wid < 0 || wid >= len(workers){
	return
    }

    w := workers[wid]
    if w == nil || w.sid == "" {
	return
    }

    for k,v := range workers {
	if v == nil || v.sid == "" {
	    continue
	}

	if k == wid {
	    continue
	}

	if strings.EqualFold(v.sid, w.sid) {
	    w.PreSaveSmpcMsg = append(w.PreSaveSmpcMsg,v.PreSaveSmpcMsg...)
	    v.bwire <-true // Release the worker thread 
	}
    }
}

func SetUpMsgList(msg string, enode string) {

	v := RecvMsg{msg: msg, sender: enode}
	//rpc-req
	rch := make(chan interface{}, 1)
	req := RPCReq{rpcdata: &v, ch: rch}
	RPCReqQueue <- req
}

func SetUpMsgList3(msg string, enode string,rch chan interface{}) {

	v := RecvMsg{msg: msg, sender: enode}
	//rpc-req
	req := RPCReq{rpcdata: &v, ch: rch}
	RPCReqQueue <- req
}

//-----------------------------------------------------------------

type WorkReq interface {
    Run(workid int, ch chan interface{}) bool
}

type RecvMsg struct {
	msg    string
	sender string
}

func (self *RecvMsg) Run(workid int, ch chan interface{}) bool {
	if workid < 0 || workid >= RPCMaxWorker {
		res2 := RpcSmpcRes{Ret: "", Tip: "smpc back-end internal error:get worker id fail", Err: fmt.Errorf("no find worker.")}
		ch <- res2
		return false
	}

	res := self.msg
	if res == "" {
		res2 := RpcSmpcRes{Ret: "", Tip: "smpc back-end internal error:get data fail in RecvMsg.Run", Err: fmt.Errorf("no find worker.")}
		ch <- res2
		return false
	}

	msgdata, errdec := DecryptMsg(res) //for SendMsgToPeer
	if errdec == nil {
		common.Debug("================RecvMsg.Run, decrypt msg success=================","res",res,"msgdata",msgdata)
		res = msgdata
	}
	
	mm := strings.Split(res, common.Sep)
	if len(mm) >= 3 {
		common.Debug("================RecvMsg.Run,begin to dis msg =================","res",res)
		//msg:  key-enode:C1:X1:X2....:Xn
		//msg:  key-enode1:NoReciv:enode2:C1
		DisMsg(res)
		return true
	}

	msgmap := make(map[string]string)
	err := json.Unmarshal([]byte(res), &msgmap)
	if err == nil {
	    if msgmap["Type"] == "SignData" {
		sd := &SignData{}
		if err = sd.UnmarshalJSON([]byte(msgmap["SignData"]));err == nil {
		    common.Debug("===============RecvMsg.Run,it is signdata===================","msgprex",sd.MsgPrex,"key",sd.Key,"pkx",sd.Pkx,"pky",sd.Pky)

		    ys := secp256k1.S256().Marshal(sd.Pkx, sd.Pky)
		    pubkeyhex := hex.EncodeToString(ys)

		    w := workers[workid]
		    w.sid = sd.Key
		    w.groupid = sd.GroupId
		    
		    w.NodeCnt = sd.NodeCnt
		    w.ThresHold = sd.ThresHold
		    
		    w.SmpcFrom = sd.SmpcFrom

		    smpcpks, _ := hex.DecodeString(pubkeyhex)
		    exsit,da := GetPubKeyData(smpcpks[:])
		    if exsit {
			    pd,ok := da.(*PubKeyData)
			    if ok {
				exsit,da2 := GetPubKeyData([]byte(pd.Key))
				if exsit {
					ac,ok := da2.(*AcceptReqAddrData)
					if ok {
					    HandleC1Data(ac,sd.Key)
					}
				}

			    }
		    }

		    childPKx := sd.Pkx
		    childPKy := sd.Pky 
		    if sd.InputCodeT != "" {
			da3 := getBip32cFromLocalDb(smpcpks[:])
			if da3 == nil {
			    res := RpcSmpcRes{Ret: "", Tip: "presign get bip32 fail", Err: fmt.Errorf("presign get bip32 fail")}
			    ch <- res
			    return false
			}
			bip32c := new(big.Int).SetBytes(da3)
			if bip32c == nil {
			    res := RpcSmpcRes{Ret: "", Tip: "presign get bip32 error", Err: fmt.Errorf("presign get bip32 error")}
			    ch <- res
			    return false
			}
			
			indexs := strings.Split(sd.InputCodeT, "/")
			TRb := bip32c.Bytes()
			childSKU1 := sd.Sku1
			for idxi := 1; idxi <len(indexs); idxi++ {
				h := hmac.New(sha512.New, TRb)
			    h.Write(childPKx.Bytes())
			    h.Write(childPKy.Bytes())
			    h.Write([]byte(indexs[idxi]))
				T := h.Sum(nil)
				TRb = T[32:]
				TL := new(big.Int).SetBytes(T[:32])

				childSKU1 = new(big.Int).Add(TL, childSKU1)
				childSKU1 = new(big.Int).Mod(childSKU1, secp256k1.S256().N)

				TLGx, TLGy := secp256k1.S256().ScalarBaseMult(TL.Bytes())
				childPKx, childPKy = secp256k1.S256().Add(TLGx, TLGy, childPKx, childPKy)
			}
		    }
		    
		    childpub := secp256k1.S256().Marshal(childPKx,childPKy)
		    childpubkeyhex := hex.EncodeToString(childpub)
		    addr,_,err := GetSmpcAddr(childpubkeyhex)
		    if err != nil {
			res := RpcSmpcRes{Ret: "", Tip: "get pubkey error", Err: fmt.Errorf("get pubkey error")}
			ch <- res
			return false
		    }
		    fmt.Printf("===================RecvMsg.Run, sign, pubkey = %v, inputcode = %v, addr = %v ===================\n",childpubkeyhex,sd.InputCodeT,addr)
     
		    var ch1 = make(chan interface{}, 1)
		    for i:=0;i < recalc_times;i++ {
			common.Debug("===============RecvMsg.Run,sign recalc===================","i",i,"msgprex",sd.MsgPrex,"key",sd.Key)
			if len(ch1) != 0 {
			    <-ch1
			}

			//w.Clear2()
			//Sign_ec2(sd.Key, sd.Save, sd.Sku1, sd.Txhash, sd.Keytype, sd.Pkx, sd.Pky, ch1, workid)
			Sign_ec3(sd.Key,sd.Txhash,sd.Keytype,sd.Save,childPKx,childPKy,ch1,workid,sd.Pre)
			common.Info("===============RecvMsg.Run, ec3 sign finish ===================","WaitMsgTimeGG20",WaitMsgTimeGG20)
			ret, _, cherr := GetChannelValue(WaitMsgTimeGG20 + 10, ch1)
			if ret != "" && cherr == nil {

			    ww, err2 := FindWorker(sd.MsgPrex)
			    if err2 != nil || ww == nil {
				res2 := RpcSmpcRes{Ret: "", Tip: "smpc back-end internal error:no find worker", Err: fmt.Errorf("no find worker")}
				ch <- res2
				return false
			    }

			    common.Info("===============RecvMsg.Run, ec3 sign success ===================","i",i,"get ret",ret,"cherr",cherr,"msgprex",sd.MsgPrex,"key",sd.Key)

			    ww.rsv.PushBack(ret)
			    res2 := RpcSmpcRes{Ret: ret, Tip: "", Err: nil}
			    ch <- res2
			    return true 
			}
			
			common.Info("===============RecvMsg.Run,ec3 sign fail===================","ret",ret,"cherr",cherr,"msgprex",sd.MsgPrex,"key",sd.Key)
		    }	
		    
		    res2 := RpcSmpcRes{Ret: "", Tip: "sign fail", Err: fmt.Errorf("sign fail")}
		    ch <- res2
		    return false 
		}
	    }
	    
	    if msgmap["Type"] == "PreSign" {
		ps := &PreSign{}
		if err = ps.UnmarshalJSON([]byte(msgmap["PreSign"]));err == nil {
		    w := workers[workid]
		    w.sid = ps.Nonce 
		    w.groupid = ps.Gid
		    w.SmpcFrom = ps.Pub
		    gcnt, _ := GetGroup(w.groupid)
		    w.NodeCnt = gcnt
		    w.ThresHold = gcnt

		    smpcpks, _ := hex.DecodeString(ps.Pub)
		    exsit,da := GetPubKeyData(smpcpks[:])
		    if !exsit {
			common.Debug("============================PreSign at RecvMsg.Run,not exist presign data===========================","pubkey",ps.Pub)
			res := RpcSmpcRes{Ret: "", Tip: "smpc back-end internal error:get presign data from db fail", Err: fmt.Errorf("get presign data from db fail")}
			ch <- res
			return false
		    }

		    pd,ok := da.(*PubKeyData)
		    if !ok {
			common.Debug("============================PreSign at RecvMsg.Run,presign data error==========================","pubkey",ps.Pub)
			res := RpcSmpcRes{Ret: "", Tip: "smpc back-end internal error:get presign data from db fail", Err: fmt.Errorf("get presign data from db fail")}
			ch <- res
			return false
		    }

		    nodecount, _ := GetGroup(pd.GroupId)
		    w.NodeCnt = nodecount

		    save := pd.Save
		    common.Debug("============================RecvMsg.Run==========================","w.SmpcFrom",w.SmpcFrom,"w.groupid",w.groupid,"w.NodeCnt",w.NodeCnt,"pd.GroupId",pd.GroupId)
		    /*msgmap := make(map[string]string)
		    err = json.Unmarshal([]byte(save), &msgmap)
		    if err != nil {
			res := RpcSmpcRes{Ret: "", Tip: "presign get local save data fail", Err: fmt.Errorf("presign get local save data fail")}
			ch <- res
			return false
		    }
		    kgsave := GetKGLocalDBSaveData(msgmap)
		    if kgsave == nil {
			res := RpcSmpcRes{Ret: "", Tip: "presign get local save data fail", Err: fmt.Errorf("presign get local save data fail")}
			ch <- res
			return false
		    }
		    sd := kgsave.Save
		    if sd.SkU1 == nil {
			res := RpcSmpcRes{Ret: "", Tip: "presign get sku1 fail", Err: fmt.Errorf("presign get sku1 fail")}
			ch <- res
			return false
		    }
		    sku1 := sd.SkU1*/
		    ///sku1
		    da2 := getSkU1FromLocalDb(smpcpks[:])
		    if da2 == nil {
			    res := RpcSmpcRes{Ret: "", Tip: "presign get sku1 fail", Err: fmt.Errorf("presign get sku1 fail")}
			    ch <- res
			    return false
		    }
		    sku1 := new(big.Int).SetBytes(da2)
		    if sku1 == nil {
			    res := RpcSmpcRes{Ret: "", Tip: "presign get sku1 fail", Err: fmt.Errorf("presign get sku1 fail")}
			    ch <- res
			    return false
		    }

		    childSKU1 := sku1
		    if ps.InputCode != "" {
			da4 := getBip32cFromLocalDb(smpcpks[:])
			if da4 == nil {
			    res := RpcSmpcRes{Ret: "", Tip: "presign get bip32 fail", Err: fmt.Errorf("presign get bip32 fail")}
			    ch <- res
			    return false
			}
			bip32c := new(big.Int).SetBytes(da4)
			if bip32c == nil {
			    res := RpcSmpcRes{Ret: "", Tip: "presign get bip32 error", Err: fmt.Errorf("presign get bip32 error")}
			ch <- res
			return false
			}
			
			smpcpub := (da.(*PubKeyData)).Pub
			smpcpkx, smpcpky := secp256k1.S256().Unmarshal(([]byte(smpcpub))[:])
			indexs := strings.Split(ps.InputCode, "/")
			TRb := bip32c.Bytes()
			childPKx := smpcpkx
			childPKy := smpcpky 
			for idxi := 1; idxi <len(indexs); idxi++ {
				h := hmac.New(sha512.New, TRb)
			    h.Write(childPKx.Bytes())
			    h.Write(childPKy.Bytes())
			    h.Write([]byte(indexs[idxi]))
				T := h.Sum(nil)
				TRb = T[32:]
				TL := new(big.Int).SetBytes(T[:32])

				childSKU1 = new(big.Int).Add(TL, childSKU1)
				childSKU1 = new(big.Int).Mod(childSKU1, secp256k1.S256().N)

				TLGx, TLGy := secp256k1.S256().ScalarBaseMult(TL.Bytes())
				childPKx, childPKy = secp256k1.S256().Add(TLGx, TLGy, childPKx, childPKy)
			}
		    }

		    exsit,da3 := GetPubKeyData([]byte(pd.Key))
		    ac,ok := da3.(*AcceptReqAddrData)
		    if ok {
			HandleC1Data(ac,w.sid)
		    }

		    var ch1 = make(chan interface{}, 1)
		    //pre := PreSign_ec3(w.sid,save,sku1,"ECDSA",ch1,workid)
		    pre := PreSign_ec3(w.sid,save,childSKU1,"EC256K1",ch1,workid)
		    if pre == nil {
			    res := RpcSmpcRes{Ret: "", Tip: "presign fail", Err: fmt.Errorf("presign fail")}
			    ch <- res
			    return false
		    }

		    pre.Key = w.sid
		    pre.Gid = w.groupid
		    pre.Used = false
		    pre.Index = ps.Index

		    err = PutPreSignData(ps.Pub,ps.InputCode,ps.Gid,ps.Index,pre)
		    if err != nil {
			res := RpcSmpcRes{Ret: "", Tip: "presign fail", Err: fmt.Errorf("presign fail")}
			ch <- res
			return false
		    }

		    res := RpcSmpcRes{Ret: "success", Tip: "", Err: nil}
		    ch <- res
		    return true
		}
	    }
	    
	    if msgmap["Type"] == "ComSignBrocastData" {
		signbrocast,err := UnCompressSignBrocastData(msgmap["ComSignBrocastData"])
		if err == nil {
		    _,_,_,txdata,err := CheckRaw(signbrocast.Raw)
		    if err == nil {
			sig,ok := txdata.(*TxDataSign)
			if ok {
			    pickdata := make([]*PickHashData,0)
			    for _,vv := range signbrocast.PickHash {
				pre := GetPreSignData(sig.PubKey,sig.InputCode,sig.GroupId,vv.PickKey)
				if pre == nil {
				    res := RpcSmpcRes{Ret: "", Tip: "dcrm back-end internal error:get pre-sign data fail", Err: fmt.Errorf("get pre-sign data fail.")}
				    ch <- res
				    return false
				}

				pd := &PickHashData{Hash:vv.Hash,Pre:pre}
				pickdata = append(pickdata,pd)
				DeletePreSignData(sig.PubKey,sig.InputCode,sig.GroupId,vv.PickKey)
			    }

			    signpick := &SignPickData{Raw:signbrocast.Raw,PickData:pickdata}
			    errtmp := InitAcceptData2(signpick,workid,self.sender,ch)
			    if errtmp == nil {
				return true
			    }

			    return false
			}
		    }
		}
	    }

	    if msgmap["Type"] == "ComSignData" {
		signpick,err := UnCompressSignData(msgmap["ComSignData"])
		if err == nil {
		    errtmp := InitAcceptData2(signpick,workid,self.sender,ch)
		    if errtmp == nil {
			return true
		    }

		    return false
		}
	    }
	}

	errtmp := InitAcceptData(res,workid,self.sender,ch)
	if errtmp == nil {
	    return true
	}

	return false 
}

func Handle(key string,c1data string) {
    w, err := FindWorker(key)
    if w == nil || err != nil {
	return
    }

    val, exist := C1Data.ReadMap(c1data)
    if exist {
	if w.DNode != nil && w.DNode.Round() != nil {
	    w.SmpcMsg <- val.(string)
	    go C1Data.DeleteMap(c1data)
	}
    }
}

func HandleKG(key string,uid *big.Int) {
    c1data := strings.ToLower(key + "-" + fmt.Sprintf("%v",uid) + "-" + "KGRound0Message")
    Handle(key,c1data)
    c1data = strings.ToLower(key + "-" + fmt.Sprintf("%v",uid) + "-" + "KGRound1Message")
    Handle(key,c1data)
}

func HandleSign(key string,uid *big.Int) {
    c1data := strings.ToLower(key + "-" + fmt.Sprintf("%v",uid) + "-" + "SignRound1Message")
    Handle(key,c1data)
    c1data = strings.ToLower(key + "-" + fmt.Sprintf("%v",uid) + "-" + "SignRound2Message")
    Handle(key,c1data)
}

func HandleC1Data(ac *AcceptReqAddrData,key string) {
    w, err := FindWorker(key)
    if w == nil || err != nil {
	return
    }

    //reshare only
    if ac == nil {
	exsit,da := GetReShareInfoData([]byte(key))
	if !exsit {
	    return
	}

	ac,ok := da.(*AcceptReShareData)
	if !ok || ac == nil {
	    return
	}

	_, enodes := GetGroup(ac.GroupId)
	nodes := strings.Split(enodes, common.Sep2)
	for _, node := range nodes {
	    node2 := ParseNode(node)
	    pk := "04" + node2 
	     h := coins.NewCryptocoinHandler("FSN")
	     if h == nil {
		continue
	     }

	    fr, err := h.PublicKeyToAddress(pk)
	    if err != nil {
		continue
	    }

	    c1data := strings.ToLower(key + "-" + fr)
	    c1, exist := C1Data.ReadMap(c1data)
	    if exist {
		DisAcceptMsg(c1.(string),w.id)
		go C1Data.DeleteMap(c1data)
	    }
	}

	return
    }
    //reshare only

    if key == "" {
	return
    }
   
    _, enodes := GetGroup(ac.GroupId)
    nodes := strings.Split(enodes, common.Sep2)
    
    for _, node := range nodes {
	node2 := ParseNode(node)
	uid := DoubleHash(node2,"EC256K1")
	HandleKG(key,uid)
	HandleSign(key,uid)
	uid = DoubleHash(node2,"ED25519")
	HandleKG(key,uid)
	HandleSign(key,uid)
    }
    
    mms := strings.Split(ac.Sigs, common.Sep)
    if len(mms) < 3 { //1:eid1:acc1
	return
    }

    count := (len(mms)-1)/2
    for j := 0;j<count;j++ {
	from := mms[2*j+2]
	c1data := strings.ToLower(key + "-" + from)
	c1, exist := C1Data.ReadMap(c1data)
	if exist {
	    DisAcceptMsg(c1.(string),w.id)
	    go C1Data.DeleteMap(c1data)
	}
    }
}

func DisAcceptMsg(raw string,workid int) {
    if raw == "" || workid < 0 || workid >= len(workers) {
	return
    }

    w := workers[workid]
    if w == nil {
	return
    }

    common.Debug("=====================DisAcceptMsg call CheckRaw================","raw ",raw)
    key,from,_,txdata,err := CheckRaw(raw)
    common.Debug("=====================DisAcceptMsg=================","key",key,"err",err)
    if err != nil {
	return
    }
    
    _,ok := txdata.(*TxDataReqAddr)
    if ok {
	if Find(w.msg_acceptreqaddrres,raw) {
	    common.Debug("======================DisAcceptMsg,the msg is reqaddr tx,and already in list.===========================","key",key,"from",from)
		return
	}

	w.msg_acceptreqaddrres.PushBack(raw)
	if w.msg_acceptreqaddrres.Len() >= w.NodeCnt {
	    if !CheckReply(w.msg_acceptreqaddrres,Rpc_REQADDR,key) {
		common.Debug("=====================DisAcceptMsg,the msg is reqaddr tx,check reply fail===================","key",key,"from",from)
		return
	    }

	    common.Debug("=====================DisAcceptMsg,the msg is reqaddr tx,check reply success and will set timeout channel===================","key",key,"from",from)
	    w.bacceptreqaddrres <- true
	    exsit,da := GetReqAddrInfoData([]byte(key))
	    if !exsit {
		return
	    }

	    ac,ok := da.(*AcceptReqAddrData)
	    if !ok || ac == nil {
		return
	    }

	    common.Debug("=====================DisAcceptMsg,the msg is reqaddr tx,set acceptReqAddrChan ===================","key",key,"from",from)
	    workers[ac.WorkId].acceptReqAddrChan <- "go on"
	}
    }
    
    sig2,ok := txdata.(*TxDataSign)
    if ok {
	    common.Debug("======================DisAcceptMsg, get the msg and it is sign tx===========================","key",key,"from",from,"raw",raw)
	if Find(w.msg_acceptsignres, raw) {
	    common.Info("======================DisAcceptMsg,the msg is sign tx,and already in list.===========================","key",key,"from",from)
	    return
	}

	    common.Debug("======================DisAcceptMsg,the msg is sign tx,and put it into list.===========================","key",key,"from",from,"sig",sig2)
	w.msg_acceptsignres.PushBack(raw)
	if w.msg_acceptsignres.Len() >= w.ThresHold {
	    if !CheckReply(w.msg_acceptsignres,Rpc_SIGN,key) {
		return
	    }

	    w.bacceptsignres <- true
	    exsit,da := GetSignInfoData([]byte(key))
	    if !exsit {
		return
	    }

	    ac,ok := da.(*AcceptSignData)
	    if !ok || ac == nil {
		return
	    }

	    workers[ac.WorkId].acceptSignChan <- "go on"
	}
    }
    
    _,ok = txdata.(*TxDataReShare)
    if ok {
	if Find(w.msg_acceptreshareres, raw) {
	    return
	}

	w.msg_acceptreshareres.PushBack(raw)
	if w.msg_acceptreshareres.Len() >= w.NodeCnt {
	    if !CheckReply(w.msg_acceptreshareres,Rpc_RESHARE,key) {
		return
	    }

	    w.bacceptreshareres <- true
	    exsit,da := GetReShareInfoData([]byte(key))
	    if !exsit {
		return
	    }

	    ac,ok := da.(*AcceptReShareData)
	    if !ok || ac == nil {
		return
	    }

	    workers[ac.WorkId].acceptReShareChan <- "go on"
	}
    }
    
    acceptreq,ok := txdata.(*TxDataAcceptReqAddr)
    if ok {
	if Find(w.msg_acceptreqaddrres,raw) {
	    common.Debug("======================DisAcceptMsg,the msg is acceptereqaddr tx,and already in list.===========================","key",acceptreq.Key,"from",from)
		return
	}

	w.msg_acceptreqaddrres.PushBack(raw)
	if w.msg_acceptreqaddrres.Len() >= w.NodeCnt {
	    if !CheckReply(w.msg_acceptreqaddrres,Rpc_REQADDR,acceptreq.Key) {
		common.Debug("=====================DisAcceptMsg,the msg is acceptereqaddr tx,check reply fail===================","key",acceptreq.Key,"from",from)
		return
	    }

	    common.Debug("=====================DisAcceptMsg,the msg is acceptereqaddr tx,check reply success and will set timeout channel===================","key",acceptreq.Key,"from",from)
	    w.bacceptreqaddrres <- true
	    exsit,da := GetReqAddrInfoData([]byte(acceptreq.Key))
	    if !exsit {
		return
	    }

	    ac,ok := da.(*AcceptReqAddrData)
	    if !ok || ac == nil {
		return
	    }

	    common.Debug("=====================DisAcceptMsg,the msg is acceptereqaddr tx,set acceptReqAddrChan ===================","key",acceptreq.Key,"from",from)
	    workers[ac.WorkId].acceptReqAddrChan <- "go on"
	}
    }
    
    acceptsig,ok := txdata.(*TxDataAcceptSign)
    if ok {
	    common.Debug("======================DisAcceptMsg, get the msg and it is accept sign tx===========================","key",acceptsig.Key,"from",from,"raw",raw)
	if Find(w.msg_acceptsignres, raw) {
	    common.Info("======================DisAcceptMsg,the msg is accept sign tx,and already in list.===========================","sig key",acceptsig.Key,"from",from)
	    return
	}

	    common.Debug("======================DisAcceptMsg,the msg is accept sign tx,and put it into list.===========================","sig key",acceptsig.Key,"from",from,"accept sig",acceptsig)
	w.msg_acceptsignres.PushBack(raw)
	if w.msg_acceptsignres.Len() >= w.ThresHold {
	    if !CheckReply(w.msg_acceptsignres,Rpc_SIGN,acceptsig.Key) {
		return
	    }

	    common.Info("======================DisAcceptMsg,the msg is accept sign tx,and check reply success and will set timeout channel.===========================","sig key",acceptsig.Key,"from",from)
	    w.bacceptsignres <- true
	    exsit,da := GetSignInfoData([]byte(acceptsig.Key))
	    if !exsit {
		return
	    }

	    ac,ok := da.(*AcceptSignData)
	    if !ok || ac == nil {
		return
	    }

	    workers[ac.WorkId].acceptSignChan <- "go on"
	}
    }
    
    acceptreshare,ok := txdata.(*TxDataAcceptReShare)
    if ok {
	if Find(w.msg_acceptreshareres, raw) {
	    return
	}

	w.msg_acceptreshareres.PushBack(raw)
	if w.msg_acceptreshareres.Len() >= w.NodeCnt {
	    if !CheckReply(w.msg_acceptreshareres,Rpc_RESHARE,acceptreshare.Key) {
		return
	    }

	    w.bacceptreshareres <- true
	    exsit,da := GetReShareInfoData([]byte(acceptreshare.Key))
	    if !exsit {
		return
	    }

	    ac,ok := da.(*AcceptReShareData)
	    if !ok || ac == nil {
		return
	    }

	    workers[ac.WorkId].acceptReShareChan <- "go on"
	}
    }
}

func IsReshareCmd(raw string) (bool,string) {
    if raw == "" {
	return false,""
    }

    key,_,_,txdata,err := CheckRaw(raw)
    if err != nil {
	return false,""
    }
    
    _,ok := txdata.(*TxDataReShare)
    if ok {
	return true,key
    }

    return false,""
}

func IsGenKeyCmd(raw string) (bool,string) {
    if raw == "" {
	return false,""
    }

    key,_,_,txdata,err := CheckRaw(raw)
    if err != nil {
	return false,""
    }
    
    _,ok := txdata.(*TxDataReqAddr)
    if ok {
	return true,key
    }

    return false,""
}

func IsPreGenSignData(raw string) (string,bool) {
    msgmap := make(map[string]string)
    err := json.Unmarshal([]byte(raw), &msgmap)
    if err == nil {
	if msgmap["Type"] == "PreSign" {
	    sd := &PreSign{}
	    if err = sd.UnmarshalJSON([]byte(msgmap["SignData"]));err == nil {
	    return sd.Nonce,true
	    }
	}
    }

    return "",false
}

func IsEDSignCmd(raw string) (string,bool) {
    if raw == "" {
	return "",false
    }

    key,_,_,txdata,err := CheckRaw(raw)
    if err != nil {
	return "",false
    }
    
    sig,ok := txdata.(*TxDataSign)
    if ok {
	if sig.Keytype == "ED25519" {
	    return key,true
	}

	return "",false
    }

    return "",false
}

func IsSignDataCmd(raw string) (string,bool) {
    msgmap := make(map[string]string)
    err := json.Unmarshal([]byte(raw), &msgmap)

    if err == nil {
	if msgmap["Type"] == "SignData" {
	    sd := &SignData{}
	    if err = sd.UnmarshalJSON([]byte(msgmap["SignData"]));err == nil {
	    return sd.Key,true
	    }
	}
    }

    return "",false
}

func InitAcceptData(raw string,workid int,sender string,ch chan interface{}) error {
    if raw == "" || workid < 0 || sender == "" {
	res := RpcSmpcRes{Ret: "", Tip: "init accept data fail.", Err: fmt.Errorf("init accept data fail")}
	ch <- res
	return fmt.Errorf("init accept data fail")
    }

    key,from,nonce,txdata,err := CheckRaw(raw)
    common.Info("=====================InitAcceptData,get result from call CheckRaw ================","key",key,"from",from,"err",err,"raw",raw)
    if err != nil {
	common.Info("===============InitAcceptData,check raw error===================","err ",err)
	res := RpcSmpcRes{Ret: "", Tip: err.Error(), Err: err}
	ch <- res
	return err
    }
    
    req,ok := txdata.(*TxDataReqAddr)
    if ok {

	common.Info("===============InitAcceptData, check reqaddr raw success==================","raw ",raw,"key ",key,"from ",from,"nonce ",nonce,"txdata ",req)
	exsit,_ := GetReqAddrInfoData([]byte(key))
	if !exsit {
	    cur_nonce, _, _ := GetReqAddrNonce(from)
	    cur_nonce_num, _ := new(big.Int).SetString(cur_nonce, 10)
	    new_nonce_num, _ := new(big.Int).SetString(nonce, 10)
	    common.Debug("===============InitAcceptData============","reqaddr cur_nonce_num ",cur_nonce_num,"reqaddr new_nonce_num ",new_nonce_num,"key ",key)
	    if new_nonce_num.Cmp(cur_nonce_num) >= 0 {
		_, err := SetReqAddrNonce(from,nonce)
		if err == nil {
		    ars := GetAllReplyFromGroup(workid,req.GroupId,Rpc_REQADDR,sender)
		    sigs,err := GetGroupSigsDataByRaw(raw) 
		    common.Info("=================InitAcceptData================","get group sigs ",sigs,"err ",err,"key ",key)
		    if err != nil {
			res := RpcSmpcRes{Ret: "", Tip: err.Error(), Err: err}
			ch <- res
			return err
		    }

		    ac := &AcceptReqAddrData{Initiator:sender,Account: from, Cointype: req.Keytype, GroupId: req.GroupId, Nonce: nonce, LimitNum: req.ThresHold, Mode: req.Mode, TimeStamp: req.TimeStamp, Deal: "false", Accept: "false", Status: "Pending", PubKey: "", Tip: "", Error: "", AllReply: ars, WorkId: workid,Sigs:sigs}
		    err = SaveAcceptReqAddrData(ac)
		    common.Info("===================call SaveAcceptReqAddrData finish====================","account ",from,"err ",err,"key ",key)
		   if err == nil {
			rch := make(chan interface{}, 1)
			w := workers[workid]
			w.sid = key 
			w.groupid = req.GroupId
			w.limitnum = req.ThresHold
			gcnt, _ := GetGroup(w.groupid)
			w.NodeCnt = gcnt
			w.ThresHold = w.NodeCnt

			nums := strings.Split(w.limitnum, "/")
			if len(nums) == 2 {
			    nodecnt, err := strconv.Atoi(nums[1])
			    if err == nil {
				w.NodeCnt = nodecnt
			    }

			    th, err := strconv.Atoi(nums[0])
			    if err == nil {
				w.ThresHold = th 
			    }
			}

			/////////////
			if req.Mode == "0" { // self-group
				////
				var reply bool
				var tip string
				timeout := make(chan bool, 1)
				go func(wid int) {
					cur_enode = discover.GetLocalID().String() //GetSelfEnode()
					agreeWaitTime := 10 * time.Minute
					agreeWaitTimeOut := time.NewTicker(agreeWaitTime)
					if wid < 0 || wid >= len(workers) || workers[wid] == nil {
						ars := GetAllReplyFromGroup(w.id,req.GroupId,Rpc_REQADDR,sender)	
						_,err = AcceptReqAddr(sender,from, req.Keytype, req.GroupId, nonce, req.ThresHold, req.Mode, "false", "false", "Failure", "", "workid error", "workid error", ars, wid,"")
						if err != nil {
						    tip = "accept reqaddr error"
						    reply = false
						    timeout <- true
						    return
						}

						tip = "worker id error"
						reply = false
						timeout <- true
						return
					}

					wtmp2 := workers[wid]
					for {
						select {
						case account := <-wtmp2.acceptReqAddrChan:
							common.Debug("(self *RecvMsg) Run(),", "account= ", account, "key = ", key)
							ars := GetAllReplyFromGroup(w.id,req.GroupId,Rpc_REQADDR,sender)
							common.Info("================== InitAcceptData,get all AcceptReqAddrRes====================","raw ",raw,"result ",ars,"key ",key)
							
							//bug
							reply = true
							for _,nr := range ars {
							    if !strings.EqualFold(nr.Status,"Agree") {
								reply = false
								break
							    }
							}
							//

							if !reply {
								tip = "don't accept req addr"
								_,err = AcceptReqAddr(sender,from, req.Keytype, req.GroupId,nonce, req.ThresHold, req.Mode, "false", "false", "Failure", "", "don't accept req addr", "don't accept req addr", ars, wid,"")
								if err != nil {
								    tip = "don't accept req addr and accept reqaddr error"
								    timeout <- true
								    return
								}
							} else {
								tip = ""
								_,err = AcceptReqAddr(sender,from, req.Keytype, req.GroupId,nonce, req.ThresHold, req.Mode, "false", "true", "Pending", "", "", "", ars, wid,"")
								if err != nil {
								    tip = "accept reqaddr error"
								    timeout <- true
								    return
								}
							}

							///////
							timeout <- true
							return
						case <-agreeWaitTimeOut.C:
							common.Info("================== InitAcceptData, agree wait timeout==================","raw ",raw,"key ",key)
							ars := GetAllReplyFromGroup(w.id,req.GroupId,Rpc_REQADDR,sender)
							//bug: if self not accept and timeout
							_,err = AcceptReqAddr(sender,from, req.Keytype, req.GroupId, nonce, req.ThresHold, req.Mode, "false", "false", "Timeout", "", "get other node accept req addr result timeout", "get other node accept req addr result timeout", ars, wid,"")
							if err != nil {
							    tip = "get other node accept req addr result timeout and accept reqaddr fail"
							    reply = false
							    timeout <- true
							    return
							}

							tip = "get other node accept req addr result timeout"
							reply = false
							//

							timeout <- true
							return
						}
					}
				}(workid)

				if len(workers[workid].acceptWaitReqAddrChan) == 0 {
					workers[workid].acceptWaitReqAddrChan <- "go on"
				}

				DisAcceptMsg(raw,workid)
				HandleC1Data(ac,key)

				<-timeout

				common.Debug("================== InitAcceptData======================","raw ",raw,"the terminal accept req addr result ",reply,"key ",key)

				ars := GetAllReplyFromGroup(w.id,req.GroupId,Rpc_REQADDR,sender)
				if !reply {
					if tip == "get other node accept req addr result timeout" {
						_,err = AcceptReqAddr(sender,from, req.Keytype, req.GroupId, nonce, req.ThresHold, req.Mode, "false", "", "Timeout", "", tip, "don't accept req addr.", ars, workid,"")
					} else {
						_,err = AcceptReqAddr(sender,from, req.Keytype, req.GroupId, nonce, req.ThresHold, req.Mode, "false", "", "Failure", "", tip, "don't accept req addr.", ars, workid,"")
					}

					if err != nil {
					    res := RpcSmpcRes{Ret:"", Tip: tip, Err: fmt.Errorf("don't accept req addr.")}
					    ch <- res
					    return fmt.Errorf("don't accept req addr.")
					}

					res := RpcSmpcRes{Ret: strconv.Itoa(workid) + common.Sep + "rpc_req_smpcaddr", Tip: tip, Err: fmt.Errorf("don't accept req addr.")}
					ch <- res
					return fmt.Errorf("don't accept req addr.")
				}
			} else {
				if len(workers[workid].acceptWaitReqAddrChan) == 0 {
					workers[workid].acceptWaitReqAddrChan <- "go on"
				}

				ars := GetAllReplyFromGroup(w.id,req.GroupId,Rpc_REQADDR,sender)
				_,err = AcceptReqAddr(sender,from, req.Keytype, req.GroupId,nonce, req.ThresHold, req.Mode, "false", "true", "Pending", "", "", "", ars, workid,"")
				if err != nil {
				    res := RpcSmpcRes{Ret:"", Tip: err.Error(), Err: err}
				    ch <- res
				    return err
				}
			}

			smpc_genPubKey(w.sid, from, req.Keytype, rch, req.Mode, nonce)
			chret, tip, cherr := GetChannelValue(waitall, rch)
			if cherr != nil {
				ars := GetAllReplyFromGroup(w.id,req.GroupId,Rpc_REQADDR,sender)
				_,err = AcceptReqAddr(sender,from, req.Keytype, req.GroupId, nonce, req.ThresHold, req.Mode, "false", "", "Failure", "", tip, cherr.Error(), ars, workid,"")
				if err != nil {
				    res := RpcSmpcRes{Ret:"", Tip:err.Error(), Err:err}
				    ch <- res
				    return err
				}

				res := RpcSmpcRes{Ret: strconv.Itoa(workid) + common.Sep + "rpc_req_smpcaddr", Tip: tip, Err: cherr}
				ch <- res
				return cherr 
			}

			res := RpcSmpcRes{Ret: strconv.Itoa(workid) + common.Sep + "rpc_req_smpcaddr" + common.Sep + chret, Tip: "", Err: nil}
			ch <- res
			return nil
		   }
		}
	    }
	}
    }

    sig,ok := txdata.(*TxDataSign)
    if ok {
	common.Debug("===============InitAcceptData, it is sign txdata and check sign raw success==================","key ",key,"from ",from,"nonce ",nonce)
	exsit,_ := GetSignInfoData([]byte(key))
	if !exsit {
	    cur_nonce, _, _ := GetSignNonce(from)
	    cur_nonce_num, _ := new(big.Int).SetString(cur_nonce, 10)
	    new_nonce_num, _ := new(big.Int).SetString(nonce, 10)
	    common.Debug("===============InitAcceptData===============","sign cur_nonce_num ",cur_nonce_num,"sign new_nonce_num ",new_nonce_num,"key ",key)
	    //if new_nonce_num.Cmp(cur_nonce_num) >= 0 {
		//_, err := SetSignNonce(from,nonce)
		_, err := SetSignNonce(from,cur_nonce) //bug
		if err == nil {
		    ars := GetAllReplyFromGroup(workid,sig.GroupId,Rpc_SIGN,sender)
		    ac := &AcceptSignData{Initiator:sender,Account: from, GroupId: sig.GroupId, Nonce: nonce, PubKey: sig.PubKey, MsgHash: sig.MsgHash, MsgContext: sig.MsgContext, Keytype: sig.Keytype, LimitNum: sig.ThresHold, Mode: sig.Mode, TimeStamp: sig.TimeStamp, Deal: "false", Accept: "false", Status: "Pending", Rsv: "", Tip: "", Error: "", AllReply: ars, WorkId:workid}
		    err = SaveAcceptSignData(ac)
		    if err == nil {
			common.Debug("===============InitAcceptData,save sign accept data finish===================","ars ",ars,"key ",key)
			w := workers[workid]
			w.sid = key 
			w.groupid = sig.GroupId 
			w.limitnum = sig.ThresHold
			gcnt, _ := GetGroup(w.groupid)
			w.NodeCnt = gcnt
			w.ThresHold = w.NodeCnt

			nums := strings.Split(w.limitnum, "/")
			if len(nums) == 2 {
			    nodecnt, err := strconv.Atoi(nums[1])
			    if err == nil {
				w.NodeCnt = nodecnt
			    }

			    w.ThresHold = gcnt
			}

			w.SmpcFrom = sig.PubKey  // pubkey replace smpcfrom in sign
			
			if sig.Mode == "0" { // self-group
				////
				var reply bool
				var tip string
				timeout := make(chan bool, 1)
				go func(wid int) {
					cur_enode = discover.GetLocalID().String() //GetSelfEnode()
					agreeWaitTime := 2 * time.Minute
					agreeWaitTimeOut := time.NewTicker(agreeWaitTime)

					wtmp2 := workers[wid]

					for {
						select {
						case account := <-wtmp2.acceptSignChan:
							common.Debug("InitAcceptData,", "account= ", account, "key = ", key)
							ars := GetAllReplyFromGroup(w.id,sig.GroupId,Rpc_SIGN,sender)
							common.Debug("================== InitAcceptData , get all AcceptSignRes===============","result ",ars,"key ",key)
							
							//TODO//
							reply = true
							recount := 0
							for _,nr := range ars {
							    if strings.EqualFold(nr.Status,"Agree") {
								    recount++
							    }
							}
							if recount < w.ThresHold {
								reply = false
							}
							////////
							//

							if !reply {
								tip = "don't accept sign"
								_,err = AcceptSign(sender,from,sig.PubKey,sig.MsgHash,sig.Keytype,sig.GroupId,nonce,sig.ThresHold,sig.Mode,"true", "false", "Failure", "", "don't accept sign", "don't accept sign", ars,wid)
							} else {
							    	common.Debug("=======================InitAcceptData,11111111111111,set sign pending=============================","key",key)
								tip = ""
								_,err = AcceptSign(sender,from,sig.PubKey,sig.MsgHash,sig.Keytype,sig.GroupId,nonce,sig.ThresHold,sig.Mode,"false", "true", "Pending", "", "", "", ars,wid)
							}

							if err != nil {
							    tip = tip + " and accept sign data fail"
							}

							///////
							timeout <- true
							return
						case <-agreeWaitTimeOut.C:
							common.Debug("================== InitAcceptData , agree wait timeout=============","key ",key)
							ars := GetAllReplyFromGroup(w.id,sig.GroupId,Rpc_SIGN,sender)
							_,err = AcceptSign(sender,from,sig.PubKey,sig.MsgHash,sig.Keytype,sig.GroupId,nonce,sig.ThresHold,sig.Mode,"true", "false", "Timeout", "", "get other node accept sign result timeout", "get other node accept sign result timeout", ars,wid)
							reply = false
							tip = "get other node accept sign result timeout"
							if err != nil {
							    tip = tip + " and accept sign data fail"
							}
							//

							timeout <- true
							return
						}
					}
				}(workid)

				if len(workers[workid].acceptWaitSignChan) == 0 {
					workers[workid].acceptWaitSignChan <- "go on"
				}

				DisAcceptMsg(raw,workid)
				common.Debug("===============InitAcceptData, call DisAcceptMsg finish===================","key ",key)
				reqaddrkey := GetReqAddrKeyByOtherKey(key,Rpc_SIGN)
				if reqaddrkey == "" {
				    res := RpcSmpcRes{Ret: "", Tip: "smpc back-end internal error:get req addr key fail", Err: fmt.Errorf("get reqaddr key fail")}
				    ch <- res
				    return fmt.Errorf("get reqaddr key fail") 
				}

				exsit,da := GetPubKeyData([]byte(reqaddrkey))
				if !exsit {
					common.Debug("===============InitAcceptData, get req addr key by other key fail ===================","key ",key)
				    res := RpcSmpcRes{Ret: "", Tip: "smpc back-end internal error:get reqaddr sigs data fail", Err: fmt.Errorf("get reqaddr sigs data fail")}
				    ch <- res
				    return fmt.Errorf("get reqaddr sigs data fail") 
				}

				acceptreqdata,ok := da.(*AcceptReqAddrData)
				if !ok || acceptreqdata == nil {
					common.Debug("===============InitAcceptData, get req addr key by other key error ===================","key ",key)
				    res := RpcSmpcRes{Ret: "", Tip: "smpc back-end internal error:get reqaddr sigs data fail", Err: fmt.Errorf("get reqaddr sigs data fail")}
				    ch <- res
				    return fmt.Errorf("get reqaddr sigs data fail") 
				}

				common.Debug("===============InitAcceptData, start call HandleC1Data===================","reqaddrkey ",reqaddrkey,"key ",key)

				HandleC1Data(acceptreqdata,key)

				<-timeout

				if !reply {
					if tip == "get other node accept sign result timeout" {
						ars := GetAllReplyFromGroup(w.id,sig.GroupId,Rpc_SIGN,sender)
						_,err = AcceptSign(sender,from,sig.PubKey,sig.MsgHash,sig.Keytype,sig.GroupId,nonce,sig.ThresHold,sig.Mode,"true", "", "Timeout", "", "get other node accept sign result timeout", "get other node accept sign result timeout", ars,workid)
					} 

					res := RpcSmpcRes{Ret:"", Tip: tip, Err: fmt.Errorf("don't accept sign.")}
					ch <- res
					return fmt.Errorf("don't accept sign.")
				}
			} else {
				if len(workers[workid].acceptWaitSignChan) == 0 {
					workers[workid].acceptWaitSignChan <- "go on"
				}

				ars := GetAllReplyFromGroup(w.id,sig.GroupId,Rpc_SIGN,sender)
				common.Debug("=======================InitAcceptData,2222222222222222222,set sign pending=============================","key",key)
				_,err = AcceptSign(sender,from,sig.PubKey,sig.MsgHash,sig.Keytype,sig.GroupId,nonce,sig.ThresHold,sig.Mode,"false", "true", "Pending", "", "","", ars,workid)
				if err != nil {
				    res := RpcSmpcRes{Ret:"", Tip: err.Error(), Err:err}
				    ch <- res
				    return err
				}
			}

			common.Debug("===============InitAcceptData,begin to sign=================","sig.MsgHash ",sig.MsgHash,"sig.Mode ",sig.Mode,"key ",key)
			rch := make(chan interface{}, 1)
			sign(w.sid, from,sig.PubKey,"",sig.MsgHash,sig.Keytype,nonce,sig.Mode,nil,rch)
			chret, tip, cherr := GetChannelValue(waitall, rch)
			common.Debug("================== InitAcceptData,finish sig.================","return sign result ",chret,"err ",cherr,"key ",key)
			if chret != "" {
				res := RpcSmpcRes{Ret: chret, Tip: "", Err: nil}
				ch <- res
				return nil
			}

			ars := GetAllReplyFromGroup(w.id,sig.GroupId,Rpc_SIGN,sender)
			if tip == "get other node accept sign result timeout" {
				_,err = AcceptSign(sender,from,sig.PubKey,sig.MsgHash,sig.Keytype,sig.GroupId,nonce,sig.ThresHold,sig.Mode,"true", "", "Timeout", "", tip,cherr.Error(),ars,workid)
			} 

			if cherr != nil {
				res := RpcSmpcRes{Ret: "", Tip: tip, Err: cherr}
				ch <- res
				return cherr
			}

			res := RpcSmpcRes{Ret: "", Tip: tip, Err: fmt.Errorf("sign fail.")}
			ch <- res
			return fmt.Errorf("sign fail.")
		    } else {
			common.Debug("===============InitAcceptData, it is sign txdata,but save accept data fail==================","key ",key,"from ",from)
		    }
		} else {
			common.Debug("===============InitAcceptData, it is sign txdata,but set nonce fail==================","key ",key,"from ",from)
		}
	    //}
	} else {
		common.Debug("===============InitAcceptData, it is sign txdata,but has handled before==================","key ",key,"from ",from)
	}
    }

    rh,ok := txdata.(*TxDataReShare)
    if ok {
	ars := GetAllReplyFromGroup(workid,rh.GroupId,Rpc_RESHARE,sender)
	sigs,err := GetGroupSigsDataByRaw(raw) 
	common.Debug("=================InitAcceptData,reshare=================","get group sigs ",sigs,"err ",err,"key ",key)
	if err != nil {
	    res := RpcSmpcRes{Ret: "", Tip: err.Error(), Err: err}
	    ch <- res
	    return err
	}

	ac := &AcceptReShareData{Initiator:sender,Account: from, GroupId: rh.GroupId, TSGroupId:rh.TSGroupId, PubKey: rh.PubKey, LimitNum: rh.ThresHold, PubAccount:rh.Account, Mode:rh.Mode, Sigs:sigs, TimeStamp: rh.TimeStamp, Deal: "false", Accept: "false", Status: "Pending", NewSk: "", Tip: "", Error: "", AllReply: ars, WorkId:workid}
	err = SaveAcceptReShareData(ac)
	common.Info("===================finish call SaveAcceptReShareData======================","err ",err,"workid ",workid,"account ",from,"group id ",rh.GroupId,"pubkey ",rh.PubKey,"threshold ",rh.ThresHold,"key ",key)
	if err == nil {
	    w := workers[workid]
	    w.sid = key 
	    w.groupid = rh.TSGroupId 
	    w.limitnum = rh.ThresHold
	    gcnt, _ := GetGroup(w.groupid)
	    w.NodeCnt = gcnt
	    w.ThresHold = w.NodeCnt

	    nums := strings.Split(w.limitnum, "/")
	    if len(nums) == 2 {
		nodecnt, err := strconv.Atoi(nums[1])
		if err == nil {
		    w.NodeCnt = nodecnt
		}

		w.ThresHold = gcnt
		if w.ThresHold == 0 {
		    th,_ := strconv.Atoi(nums[0])
		    w.ThresHold = th
		}
	    }

	    w.SmpcFrom = rh.PubKey  // pubkey replace smpcfrom in reshare 

	    var reply bool
	    var tip string
	    timeout := make(chan bool, 1)
	    go func(wid int) {
		    cur_enode = discover.GetLocalID().String() //GetSelfEnode()
		    agreeWaitTime := 10 * time.Minute
		    agreeWaitTimeOut := time.NewTicker(agreeWaitTime)

		    wtmp2 := workers[wid]

		    for {
			    select {
			    case account := <-wtmp2.acceptReShareChan:
				    common.Debug("(self *RecvMsg) Run(),", "account= ", account, "key = ", key)
				    ars := GetAllReplyFromGroup(w.id,rh.GroupId,Rpc_RESHARE,sender)
				    common.Info("================== InitAcceptData, get all AcceptReShareRes================","raw ",raw,"result ",ars,"key ",key)
				    
				    //bug
				    reply = true
				    for _,nr := range ars {
					if !strings.EqualFold(nr.Status,"Agree") {
					    reply = false
					    break
					}
				    }
				    //

				    if !reply {
					    tip = "don't accept reshare"
					    _,err = AcceptReShare(sender,from, rh.GroupId, rh.TSGroupId,rh.PubKey, rh.ThresHold,rh.Mode,"false", "false", "Failure", "", "don't accept reshare", "don't accept reshare", nil, wid)
				    } else {
					    tip = ""
					    _,err = AcceptReShare(sender,from, rh.GroupId, rh.TSGroupId,rh.PubKey, rh.ThresHold,rh.Mode,"false", "false", "pending", "", "", "", ars, wid)
				    }

				    if err != nil {
					tip = tip + " and accept reshare data fail"
				    }

				    ///////
				    timeout <- true
				    return
			    case <-agreeWaitTimeOut.C:
				    common.Info("================== InitAcceptData, agree wait timeout===================","raw ",raw,"key ",key)
				    ars := GetAllReplyFromGroup(w.id,rh.GroupId,Rpc_RESHARE,sender)
				    _,err = AcceptReShare(sender,from, rh.GroupId, rh.TSGroupId,rh.PubKey, rh.ThresHold,rh.Mode,"false", "false", "Timeout", "", "get other node accept reshare result timeout", "get other node accept reshare result timeout", ars, wid)
				    reply = false
				    tip = "get other node accept reshare result timeout"
				    if err != nil {
					tip = tip + " and accept reshare data fail"
				    }
				    //

				    timeout <- true
				    return
			    }
		    }
	    }(workid)

	    if len(workers[workid].acceptWaitReShareChan) == 0 {
		    workers[workid].acceptWaitReShareChan <- "go on"
	    }

	    DisAcceptMsg(raw,workid)
	    HandleC1Data(nil,key)
	    
	    <-timeout

	    if !reply {
		    if tip == "get other node accept reshare result timeout" {
			    ars := GetAllReplyFromGroup(workid,rh.GroupId,Rpc_RESHARE,sender)
			    _,err = AcceptReShare(sender,from, rh.GroupId, rh.TSGroupId,rh.PubKey, rh.ThresHold, rh.Mode,"false", "", "Timeout", "", "get other node accept reshare result timeout", "get other node accept reshare result timeout", ars,workid)
		    } 

		    res2 := RpcSmpcRes{Ret: "", Tip: tip, Err: fmt.Errorf("don't accept reshare.")}
		    ch <- res2
		    return fmt.Errorf("don't accept reshare.")
	    }

	    rch := make(chan interface{}, 1)
	    _reshare(w.sid,from,rh.GroupId,rh.PubKey,rh.Account,rh.Mode,sigs,rch)
	    chret, tip, cherr := GetChannelValue(ch_t, rch)
	    if chret != "" {
		    res2 := RpcSmpcRes{Ret: chret, Tip: "", Err: nil}
		    ch <- res2
		    return nil 
	    }

	    if tip == "get other node accept reshare result timeout" {
		    ars := GetAllReplyFromGroup(workid,rh.GroupId,Rpc_RESHARE,sender)
		    _,err = AcceptReShare(sender,from, rh.GroupId, rh.TSGroupId,rh.PubKey, rh.ThresHold,rh.Mode,"false", "", "Timeout", "", "get other node accept reshare result timeout", "get other node accept reshare result timeout", ars,workid)
	    } 

	    if cherr != nil {
		    res2 := RpcSmpcRes{Ret:"", Tip: tip, Err: cherr}
		    ch <- res2
		    return cherr 
	    }

	    res2 := RpcSmpcRes{Ret:"", Tip: tip, Err: fmt.Errorf("reshare fail.")}
	    ch <- res2
	    return fmt.Errorf("reshare fail.")
	}
    }

    acceptreq,ok := txdata.(*TxDataAcceptReqAddr)
    if ok {
	common.Debug("===============InitAcceptData, check accept reqaddr raw success======================","raw ",raw,"key ",acceptreq.Key,"from ",from,"txdata ",acceptreq)
	w, err := FindWorker(acceptreq.Key)
	if err != nil || w == nil {
	    c1data := strings.ToLower(acceptreq.Key + "-" + from)
	    C1Data.WriteMap(c1data,raw)
	    res := RpcSmpcRes{Ret:"Failure", Tip: "get reqaddr accept data fail from db.", Err: fmt.Errorf("get reqaddr accept data fail from db when no find worker")}
	    ch <- res
	    return fmt.Errorf("get reqaddr accept data fail from db when no find worker.")
	}

	exsit,da := GetReqAddrInfoData([]byte(acceptreq.Key))
	if !exsit {
	    res := RpcSmpcRes{Ret:"Failure", Tip: "smpc back-end internal error:get reqaddr accept data fail from db", Err: fmt.Errorf("get reqaddr accept data fail from db in init accept data")}
	    ch <- res
	    return fmt.Errorf("get reqaddr accept data fail from db in init accept data.")
	}

	ac,ok := da.(*AcceptReqAddrData)
	if !ok || ac == nil {
	    res := RpcSmpcRes{Ret:"Failure", Tip: "smpc back-end internal error:decode accept data fail", Err: fmt.Errorf("decode accept data fail")}
	    ch <- res
	    return fmt.Errorf("decode accept data fail")
	}

	status := "Pending"
	accept := "false"
	if acceptreq.Accept == "AGREE" {
		accept = "true"
	} else {
		status = "Failure"
	}

	id,_ := GetWorkerId(w)
	DisAcceptMsg(raw,id)
	HandleC1Data(ac,acceptreq.Key)

	ars := GetAllReplyFromGroup(id,ac.GroupId,Rpc_REQADDR,ac.Initiator)
	tip, err := AcceptReqAddr(ac.Initiator,ac.Account, ac.Cointype, ac.GroupId, ac.Nonce, ac.LimitNum, ac.Mode, "false", accept, status, "", "", "", ars, ac.WorkId,"")
	if err != nil {
	    res := RpcSmpcRes{Ret:"Failure", Tip: tip, Err: err}
	    ch <- res
	    return err 
	}

	res := RpcSmpcRes{Ret:"Success", Tip: "", Err: nil}
	ch <- res
	return nil
    }

    acceptsig,ok := txdata.(*TxDataAcceptSign)
    if ok {
	common.Info("===============InitAcceptData, it is acceptsign and check accept sign raw success=====================","key ",acceptsig.Key,"from ",from,"accept",acceptsig.Accept,"raw",raw)
	w, err := FindWorker(acceptsig.Key)
	if err != nil || w == nil {
		common.Info("===============InitAcceptData, it is acceptsign and no find worker=====================","key ",acceptsig.Key,"from ",from)
	    c1data := strings.ToLower(acceptsig.Key + "-" + from)
	    C1Data.WriteMap(c1data,raw)
	    res := RpcSmpcRes{Ret:"Failure", Tip: "get sign accept data fail from db when no find worker.", Err: fmt.Errorf("get sign accept data fail from db when no find worker")}
	    ch <- res
	    return fmt.Errorf("get sign accept data fail from db when no find worker.")
	}

	exsit,da := GetSignInfoData([]byte(acceptsig.Key))
	if !exsit {
		common.Info("===============InitAcceptData, it is acceptsign and get sign accept data fail from db=====================","key ",acceptsig.Key,"from ",from)
	    res := RpcSmpcRes{Ret:"Failure", Tip: "smpc back-end internal error:get sign accept data fail from db in init accept data", Err: fmt.Errorf("get sign accept data fail from db in init accept data")}
	    ch <- res
	    return fmt.Errorf("get sign accept data fail from db in init accept data.")
	}

	ac,ok := da.(*AcceptSignData)
	if !ok || ac == nil {
		common.Info("===============InitAcceptData, it is acceptsign and decode accept data fail=====================","key ",acceptsig.Key,"from ",from)
	    res := RpcSmpcRes{Ret:"Failure", Tip: "smpc back-end internal error:decode accept data fail", Err: fmt.Errorf("decode accept data fail")}
	    ch <- res
	    return fmt.Errorf("decode accept data fail")
	}

	if ac.Deal == "true" || ac.Status == "Success" || ac.Status == "Failure" || ac.Status == "Timeout" {
		common.Info("===============InitAcceptData, it is acceptsign and sign has handled before=====================","key ",acceptsig.Key,"from ",from)
	    res := RpcSmpcRes{Ret:"", Tip: "sign has handled before", Err: fmt.Errorf("sign has handled before")}
	    ch <- res
	    return fmt.Errorf("sign has handled before")
	}

	status := "Pending"
	accept := "false"
	if acceptsig.Accept == "AGREE" {
		accept = "true"
	} else {
		status = "Failure"
	}

	id,_ := GetWorkerId(w)
	DisAcceptMsg(raw,id)
	reqaddrkey := GetReqAddrKeyByOtherKey(acceptsig.Key,Rpc_SIGN)
	exsit,da = GetPubKeyData([]byte(reqaddrkey))
	if !exsit {
		common.Debug("===============InitAcceptData, it is acceptsign and get reqaddr sigs data fail=====================","key ",acceptsig.Key,"from ",from)
	    res := RpcSmpcRes{Ret: "", Tip: "smpc back-end internal error:get reqaddr sigs data fail", Err: fmt.Errorf("get reqaddr sigs data fail")}
	    ch <- res
	    return fmt.Errorf("get reqaddr sigs data fail") 
	}
	acceptreqdata,ok := da.(*AcceptReqAddrData)
	if !ok || acceptreqdata == nil {
		common.Debug("===============InitAcceptData, it is acceptsign and get reqaddr sigs data fail 2222222222 =====================","key ",acceptsig.Key,"from ",from)
	    res := RpcSmpcRes{Ret: "", Tip: "smpc back-end internal error:get reqaddr sigs data fail", Err: fmt.Errorf("get reqaddr sigs data fail")}
	    ch <- res
	    return fmt.Errorf("get reqaddr sigs data fail") 
	}

	HandleC1Data(acceptreqdata,acceptsig.Key)

	ars := GetAllReplyFromGroup(id,ac.GroupId,Rpc_SIGN,ac.Initiator)
	if ac.Deal == "true" || ac.Status == "Success" || ac.Status == "Failure" || ac.Status == "Timeout" {
		common.Info("===============InitAcceptData, it is acceptsign and sign has handled before 222222222=====================","key ",acceptsig.Key,"from ",from)
	    res := RpcSmpcRes{Ret:"", Tip: "sign has handled before", Err: fmt.Errorf("sign has handled before")}
	    ch <- res
	    return fmt.Errorf("sign has handled before")
	}

	common.Debug("=======================InitAcceptData,333333333333333333333333,set sign status =============================","status",status,"key",acceptsig.Key)
	tip, err := AcceptSign(ac.Initiator,ac.Account, ac.PubKey, ac.MsgHash, ac.Keytype, ac.GroupId, ac.Nonce,ac.LimitNum,ac.Mode,"false", accept, status, "", "", "", ars, ac.WorkId)
	if err != nil {
	    res := RpcSmpcRes{Ret:"Failure", Tip: tip, Err: err}
	    ch <- res
	    return err 
	}

	res := RpcSmpcRes{Ret:"Success", Tip: "", Err: nil}
	ch <- res
	return nil
    }

    acceptrh,ok := txdata.(*TxDataAcceptReShare)
    if ok {
	common.Debug("===============InitAcceptData, check accept reshare raw success=====================","raw ",raw,"key ",acceptrh.Key,"from ",from,"txdata ",acceptrh)
	w, err := FindWorker(acceptrh.Key)
	if err != nil || w == nil {
	    c1data := strings.ToLower(acceptrh.Key + "-" + from)
	    C1Data.WriteMap(c1data,raw)
	    res := RpcSmpcRes{Ret:"Failure", Tip: "get reshare accept data fail from db when no find worker.", Err: fmt.Errorf("get reshare accept data fail from db when no find worker")}
	    ch <- res
	    return fmt.Errorf("get reshare accept data fail from db whern no find worker.")
	}

	exsit,da := GetReShareInfoData([]byte(acceptrh.Key))
	if !exsit {
	    res := RpcSmpcRes{Ret:"Failure", Tip: "smpc back-end internal error:get reshare accept data fail from db in init accept data", Err: fmt.Errorf("get reshare accept data fail from db in init accept data")}
	    ch <- res
	    return fmt.Errorf("get reshare accept data fail from db in init accept data.")
	}

	ac,ok := da.(*AcceptReShareData)
	if !ok || ac == nil {
	    res := RpcSmpcRes{Ret:"Failure", Tip: "smpc back-end internal error:decode accept data fail", Err: fmt.Errorf("decode accept data fail")}
	    ch <- res
	    return fmt.Errorf("decode accept data fail")
	}

	status := "Pending"
	accept := "false"
	if acceptrh.Accept == "AGREE" {
		accept = "true"
	} else {
		status = "Failure"
	}

	id,_ := GetWorkerId(w)
	DisAcceptMsg(raw,id)
	HandleC1Data(nil,acceptrh.Key)

	ars := GetAllReplyFromGroup(id,ac.GroupId,Rpc_RESHARE,ac.Initiator)
	tip,err := AcceptReShare(ac.Initiator,ac.Account, ac.GroupId, ac.TSGroupId,ac.PubKey, ac.LimitNum, ac.Mode,"false", accept, status, "", "", "", ars,ac.WorkId)
	if err != nil {
	    res := RpcSmpcRes{Ret:"Failure", Tip: tip, Err: err}
	    ch <- res
	    return err 
	}

	res := RpcSmpcRes{Ret:"Success", Tip: "", Err: nil}
	ch <- res
	return nil
    }
	
    common.Debug("===============InitAcceptData, it is not sign txdata and return fail ==================","key ",key,"from ",from,"nonce ",nonce)
    res := RpcSmpcRes{Ret: "", Tip: "init accept data fail.", Err: fmt.Errorf("init accept data fail")}
    ch <- res
    return fmt.Errorf("init accept data fail")
}

//==========================================================================

func GetGroupSigsDataByRaw(raw string) (string,error) {
    if raw == "" {
	return "",fmt.Errorf("raw data empty")
    }
    
    tx := new(types.Transaction)
    raws := common.FromHex(raw)
    if err := rlp.DecodeBytes(raws, tx); err != nil {
	    return "",err
    }

    signer := types.NewEIP155Signer(big.NewInt(30400)) //
    _, err := types.Sender(signer, tx)
    if err != nil {
	return "",err
    }

    var threshold string
    var mode string
    var groupsigs string
    var groupid string

    req := TxDataReqAddr{}
    err = json.Unmarshal(tx.Data(), &req)
    if err == nil && req.TxType == "REQSMPCADDR" {
	threshold = req.ThresHold
	mode = req.Mode
	groupsigs = req.Sigs
	groupid = req.GroupId
    } else {
	rh := TxDataReShare{}
	err = json.Unmarshal(tx.Data(), &rh)
	if err == nil && rh.TxType == "RESHARE" {
	    threshold = rh.ThresHold
	    mode = rh.Mode
	    groupsigs = rh.Sigs
	    groupid = rh.GroupId
	}
    }

    if threshold == "" || mode == "" || groupid == "" {
	return "",fmt.Errorf("raw data error,it is not REQSMPCADDR tx or RESHARE tx")
    }

    if mode == "1" {
	return "",nil
    }

    if mode == "0" && groupsigs == "" {
	return "",fmt.Errorf("raw data error,must have sigs data when mode = 0")
    }

    nums := strings.Split(threshold, "/")
    nodecnt, _ := strconv.Atoi(nums[1])
    if nodecnt <= 1 {
	return "",fmt.Errorf("threshold error")
    }

    sigs := strings.Split(groupsigs,"|")
    //SigN = enode://xxxxxxxx@ip:portxxxxxxxxxxxxxxxxxxxxxx
    _, enodes := GetGroup(groupid)
    nodes := strings.Split(enodes, common.Sep2)
    if nodecnt != len(sigs) {
	fmt.Printf("============================GetGroupSigsDataByRaw,nodecnt = %v,common.Sep2 = %v,enodes = %v,groupid = %v,sigs len = %v,groupsigs = %v========================\n",nodecnt,common.Sep2,enodes,groupid,len(sigs),groupsigs)
	return "",fmt.Errorf("group sigs error")
    }

    sstmp := strconv.Itoa(nodecnt)
    for j := 0; j < nodecnt; j++ {
	en := strings.Split(sigs[j], "@")
	for _, node := range nodes {
	    node2 := ParseNode(node)
	    enId := strings.Split(en[0],"//")
	    if len(enId) < 2 {
		fmt.Printf("==========================GetGroupSigsDataByRaw,len enid = %v========================\n",len(enId))
		return "",fmt.Errorf("group sigs error")
	    }

	    if strings.EqualFold(node2, enId[1]) {
		enodesigs := []rune(sigs[j])
		if len(enodesigs) <= len(node) {
		    fmt.Printf("==========================GetGroupSigsDataByRaw,node = %v,enodesigs = %v,node len = %v,enodessigs len = %v,enodes = %v,groupsigs = %v========================\n",node,enodesigs,len(node),len(enodesigs),enodes,groupsigs)
		    return "",fmt.Errorf("group sigs error")
		}

		sig := enodesigs[len(node):]
		//sigbit, _ := hex.DecodeString(string(sig[:]))
		sigbit := common.FromHex(string(sig[:]))
		if sigbit == nil {
		    fmt.Printf("==========================GetGroupSigsDataByRaw,node = %v,enodesigs = %v,node len = %v,enodessigs len = %v,enodes = %v,groupsigs = %v,sig = %v========================\n",node,enodesigs,len(node),len(enodesigs),enodes,groupsigs,sig)
		    return "",fmt.Errorf("group sigs error")
		}

		pub,err := secp256k1.RecoverPubkey(crypto.Keccak256([]byte(node2)),sigbit)
		if err != nil {
		    fmt.Printf("==========================GetGroupSigsDataByRaw,node = %v,enodesigs = %v,node len = %v,enodessigs len = %v,enodes = %v,groupsigs = %v,sig = %v,err = %v========================\n",node,enodesigs,len(node),len(enodesigs),enodes,groupsigs,sig,err)
		    return "",err
		}
		
		h := coins.NewCryptocoinHandler("FSN")
		if h != nil {
		    pubkey := hex.EncodeToString(pub)
		    from, err := h.PublicKeyToAddress(pubkey)
		    if err != nil {
			fmt.Printf("==========================GetGroupSigsDataByRaw,node = %v,enodesigs = %v,node len = %v,enodessigs len = %v,enodes = %v,groupsigs = %v,sig = %v,err = %v, pubkey = %v========================\n",node,enodesigs,len(node),len(enodesigs),enodes,groupsigs,sig,err,pubkey)
			return "",err
		    }
		    
		    //5:eid1:acc1:eid2:acc2:eid3:acc3:eid4:acc4:eid5:acc5
		    sstmp += common.Sep
		    sstmp += node2
		    sstmp += common.Sep
		    sstmp += from
		}
	    }
	}
    }

    tmps := strings.Split(sstmp,common.Sep)
    fmt.Printf("===========================GetGroupSigsDataByRaw,sstmp = %v,common.Sep = %v,tmps len = %v,nodecnt = %v==========================\n",sstmp,common.Sep,len(tmps),nodecnt)
    if len(tmps) == (2*nodecnt + 1) {
	return sstmp,nil
    }

    return "",fmt.Errorf("group sigs error")
}

func CheckGroupEnode(gid string) bool {
    if gid == "" {
	return false
    }

    groupenode := make(map[string]bool)
    _, enodes := GetGroup(gid)
    nodes := strings.Split(enodes, common.Sep2)
    for _, node := range nodes {
	node2 := ParseNode(node)
	_, ok := groupenode[strings.ToLower(node2)]
	if ok {
	    return false
	}

	groupenode[strings.ToLower(node2)] = true
    }

    return true
}

func HandleNoReciv(key string,reqer string,ower string,datatype string,wid int) {
    w := workers[wid]
    if w == nil {
	return
    }

    var l *list.List
    switch datatype {
	case "AcceptReqAddrRes":
	    l = w.msg_acceptreqaddrres
	case "AcceptSignRes":
	    l = w.msg_acceptsignres 
	case "AcceptReShareRes":
	    l = w.msg_acceptreshareres 
	case "C1":
	    l = w.msg_c1
	case "D1":
	    l = w.msg_d1_1
	case "SHARE1":
	    l = w.msg_share1
	case "NTILDEH1H2":
	    l = w.msg_zkfact
	case "ZKUPROOF":
	    l = w.msg_zku
	case "MTAZK1PROOF":
	    l = w.msg_mtazk1proof 
	case "C11":
	    l = w.msg_c11
	case "KC":
	    l = w.msg_kc
	case "MKG":
	    l = w.msg_mkg
	case "MKW":
	    l = w.msg_mkw
	case "DELTA1":
	    l = w.msg_delta1
	case "D11":
	    l = w.msg_d11_1
	case "CommitBigVAB":
	    l = w.msg_commitbigvab
	case "ZKABPROOF":
	    l = w.msg_zkabproof
	case "CommitBigUT":
	    l = w.msg_commitbigut
	case "CommitBigUTD11":
	    l = w.msg_commitbigutd11
	case "SS1":
	    l = w.msg_ss1
	case "PaillierKey":
	    l = w.msg_paillierkey
	case "EDC11":
	    l = w.msg_edc11
	case "EDZK":
	    l = w.msg_edzk
	case "EDD11":
	    l = w.msg_edd11
	case "EDSHARE1":
	    l = w.msg_edshare1
	case "EDCFSB":
	    l = w.msg_edcfsb
	case "EDC21":
	    l = w.msg_edc21
	case "EDZKR":
	    l = w.msg_edzkr
	case "EDD21":
	    l = w.msg_edd21 
	case "EDC31":
	    l = w.msg_edc31
	case "EDD31":
	    l = w.msg_edd31
	case "EDS":
	    l = w.msg_eds 
    }
    
    if l == nil {
	return
    }

    mm := make([]string,0)
    mm = append(mm,key + "-" + ower)
    mm = append(mm,datatype)
    //mm[0] = key + "-" + ower
    //mm[1] = datatype
    var next *list.Element
    for e := l.Front(); e != nil; e = next {
	    next = e.Next()

	    if e.Value == nil {
		    continue
	    }

	    s := e.Value.(string)

	    if s == "" {
		    continue
	    }

	    tmp := strings.Split(s, common.Sep)
	    tmp2 := tmp[0:2]
	    if testEq(mm, tmp2) {
		_, enodes := GetGroup(w.groupid)
		nodes := strings.Split(enodes, common.Sep2)
		for _, node := range nodes {
		    node2 := ParseNode(node)
		    if strings.EqualFold(node2,reqer) {
			SendMsgToPeer(node,s)
			break
		    }
		}

		break
	    }
    }
}

//msg: key-enode:C1:X1:X2...:Xn
//msg: key-enode1:NoReciv:enode2:C1
func DisMsg(msg string) {

	mm := strings.Split(msg, common.Sep)
	if len(mm) < 3 {
		common.Debug("======================DisMsg, < 3 for CHECKPUBKEYSTATUS================","msg",msg,"common.Sep",common.Sep,"mm len",len(mm))
		return
	}

	mms := mm[0]
	prexs := strings.Split(mms, "-")
	if len(prexs) < 2 {
		common.Debug("======================DisMsg, < 2 for CHECKPUBKEYSTATUS================","msg",msg,"mms",mms,"prexs len",len(prexs))
		return
	}

	//msg:  hash-enode:C1:X1:X2
	w, err := FindWorker(prexs[0])
	if err != nil || w == nil {
	    mmtmp := mm[0:2]
	    ss := strings.ToLower(strings.Join(mmtmp, common.Sep))
	    common.Debug("===============DisMsg,pre-save the p2p msg=============","ss",ss,"msg",msg,"key",prexs[0])
	    C1Data.WriteMap(ss,msg)

	    return
	}

	msgCode := mm[1]
	switch msgCode {
	case "NoReciv":
		key := prexs[0]
		enode1 := prexs[1]
		enode2 := mm[2]
		datatype := mm[3]
		HandleNoReciv(key,enode1,enode2,datatype,w.id)
	case "C1":
		///bug
		if w.msg_c1.Len() >= w.NodeCnt {
			return
		}
		///
		if Find(w.msg_c1, msg) {
			return
		}

		w.msg_c1.PushBack(msg)
		common.Debug("======================DisMsg, after pushback================","w.msg_c1 len",w.msg_c1.Len(),"w.NodeCnt",w.NodeCnt,"key",prexs[0])
		if w.msg_c1.Len() == w.NodeCnt {
			common.Debug("======================DisMsg, Get All C1==================","w.msg_c1 len",w.msg_c1.Len(),"w.NodeCnt",w.NodeCnt,"key",prexs[0])
			w.bc1 <- true
		}
	case "BIP32C1":
		///bug
		if w.msg_bip32c1.Len() >= w.NodeCnt {
			return
		}
		///
		if Find(w.msg_bip32c1, msg) {
			return
		}

		w.msg_bip32c1.PushBack(msg)
		if w.msg_bip32c1.Len() == w.NodeCnt {
			w.bbip32c1 <- true
		}
	case "D1":
		///bug
		if w.msg_d1_1.Len() >= w.NodeCnt {
			return
		}
		///
		if Find(w.msg_d1_1, msg) {
			return
		}

		w.msg_d1_1.PushBack(msg)
		if w.msg_d1_1.Len() == w.NodeCnt {
			w.bd1_1 <- true
		}
	case "SHARE1":
		///bug
		if w.msg_share1.Len() >= (w.NodeCnt-1) {
			return
		}
		///
		if Find(w.msg_share1, msg) {
			return
		}

		w.msg_share1.PushBack(msg)
		if w.msg_share1.Len() == (w.NodeCnt-1) {
			w.bshare1 <- true
		}
	//case "ZKFACTPROOF":
	case "NTILDEH1H2":
		///bug
		if w.msg_zkfact.Len() >= w.NodeCnt {
			return
		}
		///
		if Find(w.msg_zkfact, msg) {
			return
		}

		w.msg_zkfact.PushBack(msg)
		if w.msg_zkfact.Len() == w.NodeCnt {
			w.bzkfact <- true
		}
	case "ZKUPROOF":
		///bug
		if w.msg_zku.Len() >= w.NodeCnt {
			return
		}
		///
		if Find(w.msg_zku, msg) {
			return
		}

		w.msg_zku.PushBack(msg)
		if w.msg_zku.Len() == w.NodeCnt {
			w.bzku <- true
		}
	case "MTAZK1PROOF":
		///bug
		if w.msg_mtazk1proof.Len() >= (w.ThresHold-1) {
			return
		}
		///
		if Find(w.msg_mtazk1proof, msg) {
			return
		}

		w.msg_mtazk1proof.PushBack(msg)
		if w.msg_mtazk1proof.Len() == (w.ThresHold-1) {
			common.Debug("=====================Get All MTAZK1PROOF====================","key",prexs[0])
			w.bmtazk1proof <- true
		}
		//sign
	case "C11":
		///bug
		if w.msg_c11.Len() >= w.ThresHold {
			return
		}
		///
		if Find(w.msg_c11, msg) {
			return
		}

		common.Debug("=====================Get C11====================","msg",msg,"key",prexs[0])
		w.msg_c11.PushBack(msg)
		if w.msg_c11.Len() == w.ThresHold {
			common.Debug("=====================Get All C11====================","key",prexs[0])
			w.bc11 <- true
		}
	case "KC":
		///bug
		if w.msg_kc.Len() >= w.ThresHold {
			return
		}
		///
		if Find(w.msg_kc, msg) {
			return
		}

		w.msg_kc.PushBack(msg)
		if w.msg_kc.Len() == w.ThresHold {
			common.Debug("=====================Get All KC====================","key",prexs[0])
			w.bkc <- true
		}
	case "MKG":
		///bug
		if w.msg_mkg.Len() >= (w.ThresHold-1) {
			return
		}
		///
		if Find(w.msg_mkg, msg) {
			return
		}

		w.msg_mkg.PushBack(msg)
		if w.msg_mkg.Len() == (w.ThresHold-1) {
			common.Debug("=====================Get All MKG====================","key",prexs[0])
			w.bmkg <- true
		}
	case "MKW":
		///bug
		if w.msg_mkw.Len() >= (w.ThresHold-1) {
			return
		}
		///
		if Find(w.msg_mkw, msg) {
			return
		}

		w.msg_mkw.PushBack(msg)
		if w.msg_mkw.Len() == (w.ThresHold-1) {
			common.Debug("=====================Get All MKW====================","key",prexs[0])
			w.bmkw <- true
		}
	case "DELTA1":
		///bug
		if w.msg_delta1.Len() >= w.ThresHold {
			return
		}
		///
		if Find(w.msg_delta1, msg) {
			return
		}

		w.msg_delta1.PushBack(msg)
		if w.msg_delta1.Len() == w.ThresHold {
			common.Debug("=====================Get All DELTA1====================","key",prexs[0])
			w.bdelta1 <- true
		}
	case "D11":
		///bug
		if w.msg_d11_1.Len() >= w.ThresHold {
			return
		}
		///
		if Find(w.msg_d11_1, msg) {
			return
		}

		w.msg_d11_1.PushBack(msg)
		if w.msg_d11_1.Len() == w.ThresHold {
			common.Debug("=====================Get All D11====================","key",prexs[0])
			w.bd11_1 <- true
		}
	case "CommitBigVAB":
		///bug
		if w.msg_commitbigvab.Len() >= w.ThresHold {
			return
		}
		///
		if Find(w.msg_commitbigvab, msg) {
			return
		}

		w.msg_commitbigvab.PushBack(msg)
		if w.msg_commitbigvab.Len() == w.ThresHold {
			common.Debug("=====================Get All CommitBigVAB====================","key",prexs[0])
			w.bcommitbigvab <- true
		}
	case "ZKABPROOF":
		///bug
		if w.msg_zkabproof.Len() >= w.ThresHold {
			return
		}
		///
		if Find(w.msg_zkabproof, msg) {
			return
		}

		w.msg_zkabproof.PushBack(msg)
		if w.msg_zkabproof.Len() == w.ThresHold {
			common.Debug("=====================Get All ZKABPROOF====================","key",prexs[0])
			w.bzkabproof <- true
		}
	case "CommitBigUT":
		///bug
		if w.msg_commitbigut.Len() >= w.ThresHold {
			return
		}
		///
		if Find(w.msg_commitbigut, msg) {
			return
		}

		w.msg_commitbigut.PushBack(msg)
		if w.msg_commitbigut.Len() == w.ThresHold {
			common.Debug("=====================Get All CommitBigUT====================","key",prexs[0])
			w.bcommitbigut <- true
		}
	case "CommitBigUTD11":
		///bug
		if w.msg_commitbigutd11.Len() >= w.ThresHold {
			return
		}
		///
		if Find(w.msg_commitbigutd11, msg) {
			return
		}

		w.msg_commitbigutd11.PushBack(msg)
		if w.msg_commitbigutd11.Len() == w.ThresHold {
			common.Debug("=====================Get All CommitBigUTD11====================","key",prexs[0])
			w.bcommitbigutd11 <- true
		}
	case "SS1":
		///bug
		if w.msg_ss1.Len() >= w.ThresHold {
			return
		}
		///
		if Find(w.msg_ss1, msg) {
			return
		}

		w.msg_ss1.PushBack(msg)
		if w.msg_ss1.Len() == w.ThresHold {
			common.Info("=====================Get All SS1====================","key",prexs[0])
			w.bss1 <- true
		}
	case "PaillierKey":
		///bug
		if w.msg_paillierkey.Len() >= w.NodeCnt {
			return
		}
		///
		if Find(w.msg_paillierkey, msg) {
			return
		}

		w.msg_paillierkey.PushBack(msg)
		//if w.msg_paillierkey.Len() == w.ThresHold {
		if w.msg_paillierkey.Len() == w.NodeCnt {
			common.Debug("=====================Get All PaillierKey====================","key",prexs[0])
			w.bpaillierkey <- true
		}


	//////////////////ed
	case "EDC11":
		///bug
		if w.msg_edc11.Len() >= w.NodeCnt {
			return
		}
		///
		if Find(w.msg_edc11, msg) {
			return
		}

		w.msg_edc11.PushBack(msg)
		if w.msg_edc11.Len() == w.NodeCnt {
			w.bedc11 <- true
		}
	case "EDZK":
		///bug
		if w.msg_edzk.Len() >= w.NodeCnt {
			return
		}
		///
		if Find(w.msg_edzk, msg) {
			return
		}

		w.msg_edzk.PushBack(msg)
		if w.msg_edzk.Len() == w.NodeCnt {
			w.bedzk <- true
		}
	case "EDD11":
		///bug
		if w.msg_edd11.Len() >= w.NodeCnt {
			return
		}
		///
		if Find(w.msg_edd11, msg) {
			return
		}

		w.msg_edd11.PushBack(msg)
		if w.msg_edd11.Len() == w.NodeCnt {
			w.bedd11 <- true
		}
	case "EDSHARE1":
		///bug
		if w.msg_edshare1.Len() >= (w.NodeCnt-1) {
			return
		}
		///
		if Find(w.msg_edshare1, msg) {
			return
		}

		w.msg_edshare1.PushBack(msg)
		if w.msg_edshare1.Len() == (w.NodeCnt-1) {
			w.bedshare1 <- true
		}
	case "EDCFSB":
		///bug
		if w.msg_edcfsb.Len() >= w.NodeCnt {
			return
		}
		///
		if Find(w.msg_edcfsb, msg) {
			return
		}

		w.msg_edcfsb.PushBack(msg)
		if w.msg_edcfsb.Len() == w.NodeCnt {
			w.bedcfsb <- true
		}
	case "EDC21":
		///bug
		if w.msg_edc21.Len() >= w.NodeCnt {
			return
		}
		///
		if Find(w.msg_edc21, msg) {
			return
		}

		w.msg_edc21.PushBack(msg)
		if w.msg_edc21.Len() == w.NodeCnt {
			w.bedc21 <- true
		}
	case "EDZKR":
		///bug
		if w.msg_edzkr.Len() >= w.NodeCnt {
			return
		}
		///
		if Find(w.msg_edzkr, msg) {
			return
		}

		w.msg_edzkr.PushBack(msg)
		if w.msg_edzkr.Len() == w.NodeCnt {
			w.bedzkr <- true
		}
	case "EDD21":
		///bug
		if w.msg_edd21.Len() >= w.NodeCnt {
			return
		}
		///
		if Find(w.msg_edd21, msg) {
			return
		}

		w.msg_edd21.PushBack(msg)
		if w.msg_edd21.Len() == w.NodeCnt {
			w.bedd21 <- true
		}
	case "EDC31":
		///bug
		if w.msg_edc31.Len() >= w.NodeCnt {
			return
		}
		///
		if Find(w.msg_edc31, msg) {
			return
		}

		w.msg_edc31.PushBack(msg)
		if w.msg_edc31.Len() == w.NodeCnt {
			w.bedc31 <- true
		}
	case "EDD31":
		///bug
		if w.msg_edd31.Len() >= w.NodeCnt {
			return
		}
		///
		if Find(w.msg_edd31, msg) {
			return
		}

		w.msg_edd31.PushBack(msg)
		if w.msg_edd31.Len() == w.NodeCnt {
			w.bedd31 <- true
		}
	case "EDS":
		///bug
		if w.msg_eds.Len() >= w.NodeCnt {
			return
		}
		///
		if Find(w.msg_eds, msg) {
			return
		}

		w.msg_eds.PushBack(msg)
		if w.msg_eds.Len() == w.NodeCnt {
			w.beds <- true
		}
		///////////////////
	default:
		fmt.Println("unkown msg code")
	}
}

//==========================================================================

func Find(l *list.List, msg string) bool {
	if l == nil || msg == "" {
		return false
	}

	var next *list.Element
	for e := l.Front(); e != nil; e = next {
		next = e.Next()

		if e.Value == nil {
			continue
		}

		s := e.Value.(string)

		if s == "" {
			continue
		}

		if strings.EqualFold(s, msg) {
			return true
		}
	}

	return false
}

func testEq(a, b []string) bool {
    // If one is nil, the other must also be nil.
    if (a == nil) != (b == nil) {
        return false;
    }

    if len(a) != len(b) {
        return false
    }

    for i := range a {
	if !strings.EqualFold(a[i],b[i]) {
            return false
        }
    }

    return true
}

