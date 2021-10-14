package ec2

import (
	"encoding/json"
	"fmt"
	"math/big"
)

type NtildeH1H2 struct {
	Ntilde *big.Int
	H1     *big.Int
	H2     *big.Int
}

// GenerateNtildeH1H2 create ntilde data
func GenerateNtildeH1H2(length int) (*NtildeH1H2, *big.Int, *big.Int, *big.Int, *big.Int) {

	sp1 := <-SafePrimeCh
	sp2 := <-SafePrimeCh

	if sp1.p == nil || sp2.p == nil {
		return nil, nil, nil, nil, nil
	}

	SafePrimeCh <- sp1
	SafePrimeCh <- sp2

	NTildei := new(big.Int).Mul(sp1.P(), sp2.P())
	modNTildeI := ModInt(NTildei)

	modPQ := ModInt(new(big.Int).Mul(sp1.Q(), sp2.Q()))
	f1 := GetRandomPositiveRelativelyPrimeInt(NTildei)
	alpha := GetRandomPositiveRelativelyPrimeInt(NTildei)
	beta := modPQ.Inverse(alpha)

	h1i := modNTildeI.Mul(f1, f1)
	h2i := modNTildeI.Exp(h1i, alpha)

	ntildeH1H2 := &NtildeH1H2{Ntilde: NTildei, H1: h1i, H2: h2i}

	return ntildeH1H2, alpha, beta, sp1.Q(), sp2.Q()
}

//--------------------------------------------------------------------------

func (ntilde *NtildeH1H2) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Ntilde string `json:"Ntilde"`
		H1     string `json:"H1"`
		H2     string `json:"H2"`
	}{
		Ntilde: fmt.Sprintf("%v", ntilde.Ntilde),
		H1:     fmt.Sprintf("%v", ntilde.H1),
		H2:     fmt.Sprintf("%v", ntilde.H2),
	})
}

func (ntilde *NtildeH1H2) UnmarshalJSON(raw []byte) error {
	var nti struct {
		Ntilde string `json:"Ntilde"`
		H1     string `json:"H1"`
		H2     string `json:"H2"`
	}
	if err := json.Unmarshal(raw, &nti); err != nil {
		return err
	}

	ntilde.Ntilde, _ = new(big.Int).SetString(nti.Ntilde, 10)
	ntilde.H1, _ = new(big.Int).SetString(nti.H1, 10)
	ntilde.H2, _ = new(big.Int).SetString(nti.H2, 10)
	return nil
}

//----------------------------------------------------------------------

func CreateNt(length int) (*NtildeH1H2, *big.Int, *big.Int, *big.Int, *big.Int) {

	p, P := GetRandomPrime()
	q, Q := GetRandomPrime()

	if p == nil || q == nil || P == nil || Q == nil {
		return nil, nil, nil, nil, nil
	}

	NTildei := new(big.Int).Mul(P, Q)
	modNTildeI := ModInt(NTildei)

	modPQ := ModInt(new(big.Int).Mul(p, q))
	f1 := GetRandomPositiveRelativelyPrimeInt(NTildei)
	alpha := GetRandomPositiveRelativelyPrimeInt(NTildei)
	beta := modPQ.Inverse(alpha)
	h1i := modNTildeI.Mul(f1, f1)
	h2i := modNTildeI.Exp(h1i, alpha)

	ntildeH1H2 := &NtildeH1H2{Ntilde: NTildei, H1: h1i, H2: h2i}
	return ntildeH1H2, alpha, beta, p, q
}
