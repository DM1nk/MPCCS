
/*
 *  Copyright (C) 2018-2019  Fusion Foundation Ltd. All rights reserved.
 *  Copyright (C) 2018-2019  haijun.cai@anyswap.exchange
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
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"
	"bytes"

	"github.com/fsn-dev/cryptoCoins/coins/types"
	"github.com/anyswap/Anyswap-MPCNode/internal/common"
	"github.com/fsn-dev/cryptoCoins/tools/rlp"
	"encoding/gob"
	"compress/zlib"
	"github.com/anyswap/Anyswap-MPCNode/crypto/sha3"
	"io"
	"github.com/anyswap/Anyswap-MPCNode/internal/common/hexutil"
	"github.com/anyswap/Anyswap-MPCNode/smpc-lib/crypto/ed"
	smpclib "github.com/anyswap/Anyswap-MPCNode/smpc-lib/smpc"
	"sort"
	"container/list"
)

//---------------------------------------------------------------------------

type SmpcReq interface {
    GetReplyFromGroup(wid int,gid string,initiator string) []NodeReply 
    GetReqAddrKeyByKey(key string) string
    GetRawReply(ret *common.SafeMap,reply *RawReply)
    CheckReply(ac *AcceptReqAddrData,l *list.List,key string) bool
    DoReq(raw string,workid int,sender string,ch chan interface{}) bool
    GetGroupSigs(txdata []byte) (string,string,string,string)
    CheckTxData(txdata []byte,from string,nonce uint64) (string,string,string,interface{},error)
    DisAcceptMsg(raw string,workid int,key string)
}

//-----------------------------------------------------------------------

type RpcSmpcRes struct {
	Ret string
	Tip string
	Err error
}

func GetChannelValue(t int, obj interface{}) (string, string, error) {
	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(time.Duration(t) * time.Second) //1000 == 1s
		timeout <- true
	}()

	switch ch := obj.(type) {
	case chan interface{}:
		select {
		case v := <-ch:
			ret, ok := v.(RpcSmpcRes)
			if ok {
				return ret.Ret, ret.Tip, ret.Err
			}
		case <-timeout:
			return "", "smpc back-end internal error:get result from channel timeout", fmt.Errorf("get result from channel timeout.")
		}
	case chan string:
		select {
		case v := <-ch:
			return v, "", nil
		case <-timeout:
			return "", "smpc back-end internal error:get result from channel timeout", fmt.Errorf("get result from channel timeout.")
		}
	case chan int64:
		select {
		case v := <-ch:
			return strconv.Itoa(int(v)), "", nil
		case <-timeout:
			return "", "smpc back-end internal error:get result from channel timeout", fmt.Errorf("get result from channel timeout.")
		}
	case chan int:
		select {
		case v := <-ch:
			return strconv.Itoa(v), "", nil
		case <-timeout:
			return "", "smpc back-end internal error:get result from channel timeout", fmt.Errorf("get result from channel timeout.")
		}
	case chan bool:
		select {
		case v := <-ch:
			if !v {
				return "false", "", nil
			} else {
				return "true", "", nil
			}
		case <-timeout:
			return "", "smpc back-end internal error:get result from channel timeout", fmt.Errorf("get result from channel timeout.")
		}
	default:
		return "", "smpc back-end internal error:unknown channel type", fmt.Errorf("unknown channel type.")
	}

	return "", "smpc back-end internal error:unknown error.", fmt.Errorf("get result from channel fail,unsupported channel type.")
}

//----------------------------------------------------------------------------------------

func Encode2(obj interface{}) (string, error) {
    switch ch := obj.(type) {
	case *PubKeyData:
		var buff bytes.Buffer
		enc := gob.NewEncoder(&buff)

		err1 := enc.Encode(ch)
		if err1 != nil {
			return "", err1
		}
		return buff.String(), nil
	case *AcceptReqAddrData:
		ret,err := json.Marshal(ch)
		if err != nil {
		    return "",err
		}
		return string(ret),nil
	case *AcceptSignData:
		var buff bytes.Buffer
		enc := gob.NewEncoder(&buff)

		err1 := enc.Encode(ch)
		if err1 != nil {
		    return "", err1
		}
		return buff.String(), nil
	case *AcceptReShareData:

		var buff bytes.Buffer
		enc := gob.NewEncoder(&buff)

		err1 := enc.Encode(ch)
		if err1 != nil {
		    return "", err1
		}
		return buff.String(), nil
	default:
		return "", fmt.Errorf("encode fail.")
	}
}

func Decode2(s string, datatype string) (interface{}, error) {

	if datatype == "PubKeyData" {
		var data bytes.Buffer
		data.Write([]byte(s))

		dec := gob.NewDecoder(&data)

		var res PubKeyData
		err := dec.Decode(&res)
		if err != nil {
			return nil, err
		}

		return &res, nil
	}

	if datatype == "AcceptReqAddrData" {
		var m AcceptReqAddrData
		err := json.Unmarshal([]byte(s), &m)
		if err != nil {
		    return nil,err
		}

		return &m,nil
	}

	if datatype == "AcceptSignData" {
		var data bytes.Buffer
		data.Write([]byte(s))

		dec := gob.NewDecoder(&data)

		var res AcceptSignData
		err := dec.Decode(&res)
		if err != nil {
			return nil, err
		}

		return &res, nil
	}

	if datatype == "AcceptReShareData" {
		var data bytes.Buffer
		data.Write([]byte(s))

		dec := gob.NewDecoder(&data)

		var res AcceptReShareData
		err := dec.Decode(&res)
		if err != nil {
			return nil, err
		}

		return &res, nil
	}

	return nil, fmt.Errorf("decode fail.")
}

//--------------------------------------------------------------------------------------

func Compress(c []byte) (string, error) {
	if c == nil {
		return "", fmt.Errorf("compress fail.")
	}

	var in bytes.Buffer
	w, err := zlib.NewWriterLevel(&in, zlib.BestCompression-1)
	if err != nil {
		return "", err
	}

	_,err = w.Write(c)
	if err != nil {
	    return "",err
	}

	w.Close()

	s := in.String()
	return s, nil
}

func UnCompress(s string) (string, error) {

	if s == "" {
		return "", fmt.Errorf("param error.")
	}

	var data bytes.Buffer
	data.Write([]byte(s))

	r, err := zlib.NewReader(&data)
	if err != nil {
		return "", err
	}

	var out bytes.Buffer
	_,err = io.Copy(&out, r)
	if err != nil {
	    return "",err
	}

	return out.String(), nil
}

//---------------------------------------------------------------------------------------

type SmpcHash [32]byte

func (h SmpcHash) Hex() string { return hexutil.Encode(h[:]) }

// Keccak256Hash calculates and returns the Keccak256 hash of the input data,
// converting it to an internal Hash data structure.
func Keccak256Hash(data ...[]byte) (h SmpcHash) {
	d := sha3.NewKeccak256()
	for _, b := range data {
	    _,err := d.Write(b)
	    if err != nil {
		return h 
	    }
	}
	d.Sum(h[:0])
	return h
}

//----------------------------------------------------------------------------------------------

func DoubleHash(id string, keytype string) *big.Int {
	// Generate the random num
	// First, hash with the keccak256
	keccak256 := sha3.NewKeccak256()
	_,err := keccak256.Write([]byte(id))
	if err != nil {
	    return nil
	}

	digestKeccak256 := keccak256.Sum(nil)

	//second, hash with the SHA3-256
	sha3256 := sha3.New256()

	_,err = sha3256.Write(digestKeccak256)
	if err != nil {
	    return nil
	}

	if keytype == "ED25519" {
	    var digest [32]byte
	    copy(digest[:], sha3256.Sum(nil))

	    //////
	    var zero [32]byte
	    var one [32]byte
	    one[0] = 1
	    ed.ScMulAdd(&digest, &digest, &one, &zero)
	    //////
	    digestBigInt := new(big.Int).SetBytes(digest[:])
	    return digestBigInt
	}

	digest := sha3256.Sum(nil)
	// convert the hash ([]byte) to big.Int
	digestBigInt := new(big.Int).SetBytes(digest)
	return digestBigInt
}

func GetIds(keytype string, groupid string) smpclib.SortableIDSSlice {
	var ids smpclib.SortableIDSSlice
	_, nodes := GetGroup(groupid)
	others := strings.Split(nodes, common.Sep2)
	for _, v := range others {
		node2 := ParseNode(v) //bug??
		uid := DoubleHash(node2, keytype)
		ids = append(ids, uid)
	}
	sort.Sort(ids)
	return ids
}

//------------------------------------------------------------------------------

func GetTxTypeFromData(txdata []byte) string {
    if txdata == nil {
	return ""
    }
    
    req := TxDataReqAddr{}
    err := json.Unmarshal(txdata, &req)
    if err == nil && req.TxType == "REQSMPCADDR" {
	return "REQSMPCADDR"
    }
    
    sig := TxDataSign{}
    err = json.Unmarshal(txdata, &sig)
    if err == nil && sig.TxType == "SIGN" {
	return "SIGN"
    }

    pre := TxDataPreSignData{}
    err = json.Unmarshal(txdata, &pre)
    if err == nil && pre.TxType == "PRESIGNDATA" {
	return "PRESIGNDATA"
    }

    rh := TxDataReShare{}
    err = json.Unmarshal(txdata, &rh)
    if err == nil && rh.TxType == "RESHARE" {
	return "RESHARE"
    }

    acceptreq := TxDataAcceptReqAddr{}
    err = json.Unmarshal(txdata, &acceptreq)
    if err == nil && acceptreq.TxType == "ACCEPTREQADDR" {
	return "ACCEPTREQADDR"
    }

    acceptsig := TxDataAcceptSign{}
    err = json.Unmarshal(txdata, &acceptsig)
    if err == nil && acceptsig.TxType == "ACCEPTSIGN" {
	return "ACCEPTSIGN"
    }

    acceptrh := TxDataAcceptReShare{}
    err = json.Unmarshal(txdata, &acceptrh)
    if err == nil && acceptrh.TxType == "ACCEPTRESHARE" {
	return "ACCEPTRESHARE"
    }

    return ""
}

func CheckRaw(raw string) (string,string,string,interface{},error) {
    if raw == "" {
	return "","","",nil,fmt.Errorf("raw data empty")
    }
    
    tx := new(types.Transaction)
    raws := common.FromHex(raw)
    if err := rlp.DecodeBytes(raws, tx); err != nil {
	    return "","","",nil,err
    }

    signer := types.NewEIP155Signer(big.NewInt(30400)) //
    from, err := types.Sender(signer, tx)
    if err != nil {
	return "", "","",nil,err
    }

    var smpc_req SmpcReq
    txtype := GetTxTypeFromData(tx.Data())
    switch txtype {
	case "REQSMPCADDR":
	    smpc_req = &ReqSmpcAddr{}
	case "SIGN":
	    smpc_req = &ReqSmpcSign{}
	case "PRESIGNDATA":
	    smpc_req = &ReqSmpcSign{}
	case "RESHARE":
	    smpc_req = &ReqSmpcReshare{}
	case "ACCEPTREQADDR":
	    smpc_req = &ReqSmpcAddr{}
	case "ACCEPTSIGN": 
	    smpc_req = &ReqSmpcSign{}
	case "ACCEPTRESHARE": 
	    smpc_req = &ReqSmpcReshare{}
	default:
	    return "","","",nil,fmt.Errorf("Unsupported request type")
    }

    return smpc_req.CheckTxData(tx.Data(),from.Hex(),tx.Nonce())
}


