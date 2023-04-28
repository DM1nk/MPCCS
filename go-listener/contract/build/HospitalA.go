// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package hospitala

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// HospitalaMetaData contains all meta data concerning the Hospitala contract.
var HospitalaMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"x\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"a\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"b\",\"type\":\"uint256\"}],\"name\":\"between\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"a\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"b\",\"type\":\"uint256\"}],\"name\":\"max\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"a\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"b\",\"type\":\"uint256\"}],\"name\":\"min\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"pure\",\"type\":\"function\"}]",
	Bin: "0x610293610053600b82828239805160001a607314610046577f4e487b7100000000000000000000000000000000000000000000000000000000600052600060045260246000fd5b30600052607381538281f3fe730000000000000000000000000000000000000000301460806040526004361061004b5760003560e01c80636d5433e6146100505780637625ea10146100805780637ae2b5c7146100b0575b600080fd5b61006a6004803603810190610065919061017a565b6100e0565b60405161007791906101c9565b60405180910390f35b61009a600480360381019061009591906101e4565b6100f9565b6040516100a791906101c9565b60405180910390f35b6100ca60048036038101906100c5919061017a565b610126565b6040516100d791906101c9565b60405180910390f35b60008183116100ef57816100f1565b825b905092915050565b60008284101561010b5782905061011f565b8184111561011b5781905061011f565b8390505b9392505050565b60008183106101355781610137565b825b905092915050565b600080fd5b6000819050919050565b61015781610144565b811461016257600080fd5b50565b6000813590506101748161014e565b92915050565b600080604083850312156101915761019061013f565b5b600061019f85828601610165565b92505060206101b085828601610165565b9150509250929050565b6101c381610144565b82525050565b60006020820190506101de60008301846101ba565b92915050565b6000806000606084860312156101fd576101fc61013f565b5b600061020b86828701610165565b935050602061021c86828701610165565b925050604061022d86828701610165565b915050925092509256fea2646970667358221220553aab2dd8ab3d36c16a05dd027c3f9adf3f80416586c021d135539db22232b664736f6c637828302e382e32302d646576656c6f702e323032332e342e32362b636f6d6d69742e31346332356333380059",
}

// HospitalaABI is the input ABI used to generate the binding from.
// Deprecated: Use HospitalaMetaData.ABI instead.
var HospitalaABI = HospitalaMetaData.ABI

// HospitalaBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use HospitalaMetaData.Bin instead.
var HospitalaBin = HospitalaMetaData.Bin

// DeployHospitala deploys a new Ethereum contract, binding an instance of Hospitala to it.
func DeployHospitala(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *Hospitala, error) {
	parsed, err := HospitalaMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(HospitalaBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Hospitala{HospitalaCaller: HospitalaCaller{contract: contract}, HospitalaTransactor: HospitalaTransactor{contract: contract}, HospitalaFilterer: HospitalaFilterer{contract: contract}}, nil
}

// Hospitala is an auto generated Go binding around an Ethereum contract.
type Hospitala struct {
	HospitalaCaller     // Read-only binding to the contract
	HospitalaTransactor // Write-only binding to the contract
	HospitalaFilterer   // Log filterer for contract events
}

// HospitalaCaller is an auto generated read-only Go binding around an Ethereum contract.
type HospitalaCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// HospitalaTransactor is an auto generated write-only Go binding around an Ethereum contract.
type HospitalaTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// HospitalaFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type HospitalaFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// HospitalaSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type HospitalaSession struct {
	Contract     *Hospitala        // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// HospitalaCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type HospitalaCallerSession struct {
	Contract *HospitalaCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts    // Call options to use throughout this session
}

// HospitalaTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type HospitalaTransactorSession struct {
	Contract     *HospitalaTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts    // Transaction auth options to use throughout this session
}

// HospitalaRaw is an auto generated low-level Go binding around an Ethereum contract.
type HospitalaRaw struct {
	Contract *Hospitala // Generic contract binding to access the raw methods on
}

// HospitalaCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type HospitalaCallerRaw struct {
	Contract *HospitalaCaller // Generic read-only contract binding to access the raw methods on
}

// HospitalaTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type HospitalaTransactorRaw struct {
	Contract *HospitalaTransactor // Generic write-only contract binding to access the raw methods on
}

// NewHospitala creates a new instance of Hospitala, bound to a specific deployed contract.
func NewHospitala(address common.Address, backend bind.ContractBackend) (*Hospitala, error) {
	contract, err := bindHospitala(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Hospitala{HospitalaCaller: HospitalaCaller{contract: contract}, HospitalaTransactor: HospitalaTransactor{contract: contract}, HospitalaFilterer: HospitalaFilterer{contract: contract}}, nil
}

// NewHospitalaCaller creates a new read-only instance of Hospitala, bound to a specific deployed contract.
func NewHospitalaCaller(address common.Address, caller bind.ContractCaller) (*HospitalaCaller, error) {
	contract, err := bindHospitala(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &HospitalaCaller{contract: contract}, nil
}

// NewHospitalaTransactor creates a new write-only instance of Hospitala, bound to a specific deployed contract.
func NewHospitalaTransactor(address common.Address, transactor bind.ContractTransactor) (*HospitalaTransactor, error) {
	contract, err := bindHospitala(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &HospitalaTransactor{contract: contract}, nil
}

// NewHospitalaFilterer creates a new log filterer instance of Hospitala, bound to a specific deployed contract.
func NewHospitalaFilterer(address common.Address, filterer bind.ContractFilterer) (*HospitalaFilterer, error) {
	contract, err := bindHospitala(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &HospitalaFilterer{contract: contract}, nil
}

// bindHospitala binds a generic wrapper to an already deployed contract.
func bindHospitala(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := HospitalaMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Hospitala *HospitalaRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Hospitala.Contract.HospitalaCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Hospitala *HospitalaRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Hospitala.Contract.HospitalaTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Hospitala *HospitalaRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Hospitala.Contract.HospitalaTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Hospitala *HospitalaCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Hospitala.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Hospitala *HospitalaTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Hospitala.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Hospitala *HospitalaTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Hospitala.Contract.contract.Transact(opts, method, params...)
}

// Between is a free data retrieval call binding the contract method 0x7625ea10.
//
// Solidity: function between(uint256 x, uint256 a, uint256 b) pure returns(uint256)
func (_Hospitala *HospitalaCaller) Between(opts *bind.CallOpts, x *big.Int, a *big.Int, b *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _Hospitala.contract.Call(opts, &out, "between", x, a, b)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Between is a free data retrieval call binding the contract method 0x7625ea10.
//
// Solidity: function between(uint256 x, uint256 a, uint256 b) pure returns(uint256)
func (_Hospitala *HospitalaSession) Between(x *big.Int, a *big.Int, b *big.Int) (*big.Int, error) {
	return _Hospitala.Contract.Between(&_Hospitala.CallOpts, x, a, b)
}

// Between is a free data retrieval call binding the contract method 0x7625ea10.
//
// Solidity: function between(uint256 x, uint256 a, uint256 b) pure returns(uint256)
func (_Hospitala *HospitalaCallerSession) Between(x *big.Int, a *big.Int, b *big.Int) (*big.Int, error) {
	return _Hospitala.Contract.Between(&_Hospitala.CallOpts, x, a, b)
}

// Max is a free data retrieval call binding the contract method 0x6d5433e6.
//
// Solidity: function max(uint256 a, uint256 b) pure returns(uint256)
func (_Hospitala *HospitalaCaller) Max(opts *bind.CallOpts, a *big.Int, b *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _Hospitala.contract.Call(opts, &out, "max", a, b)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Max is a free data retrieval call binding the contract method 0x6d5433e6.
//
// Solidity: function max(uint256 a, uint256 b) pure returns(uint256)
func (_Hospitala *HospitalaSession) Max(a *big.Int, b *big.Int) (*big.Int, error) {
	return _Hospitala.Contract.Max(&_Hospitala.CallOpts, a, b)
}

// Max is a free data retrieval call binding the contract method 0x6d5433e6.
//
// Solidity: function max(uint256 a, uint256 b) pure returns(uint256)
func (_Hospitala *HospitalaCallerSession) Max(a *big.Int, b *big.Int) (*big.Int, error) {
	return _Hospitala.Contract.Max(&_Hospitala.CallOpts, a, b)
}

// Min is a free data retrieval call binding the contract method 0x7ae2b5c7.
//
// Solidity: function min(uint256 a, uint256 b) pure returns(uint256)
func (_Hospitala *HospitalaCaller) Min(opts *bind.CallOpts, a *big.Int, b *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _Hospitala.contract.Call(opts, &out, "min", a, b)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Min is a free data retrieval call binding the contract method 0x7ae2b5c7.
//
// Solidity: function min(uint256 a, uint256 b) pure returns(uint256)
func (_Hospitala *HospitalaSession) Min(a *big.Int, b *big.Int) (*big.Int, error) {
	return _Hospitala.Contract.Min(&_Hospitala.CallOpts, a, b)
}

// Min is a free data retrieval call binding the contract method 0x7ae2b5c7.
//
// Solidity: function min(uint256 a, uint256 b) pure returns(uint256)
func (_Hospitala *HospitalaCallerSession) Min(a *big.Int, b *big.Int) (*big.Int, error) {
	return _Hospitala.Contract.Min(&_Hospitala.CallOpts, a, b)
}
