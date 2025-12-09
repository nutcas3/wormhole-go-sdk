package types

import "fmt"

type TokenID struct {
	Chain   Chain
	Address string
}

func NewTokenID(chain Chain, address string) TokenID {
	return TokenID{
		Chain:   chain,
		Address: address,
	}
}

func (t TokenID) String() string {
	return fmt.Sprintf("%s:%s", t.Chain, t.Address)
}

func (t TokenID) IsNative() bool {
	return t.Address == "native"
}

type Amount struct {
	Value    string // String representation to avoid precision loss
	Decimals uint8
}

func NewAmount(value string, decimals uint8) Amount {
	return Amount{
		Value:    value,
		Decimals: decimals,
	}
}
