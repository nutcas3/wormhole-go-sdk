package types

type Network string

const (
	Mainnet Network = "Mainnet"
	Testnet Network = "Testnet"
	Devnet Network = "Devnet"
)

func (n Network) String() string {
	return string(n)
}

func (n Network) IsValid() bool {
	switch n {
	case Mainnet, Testnet, Devnet:
		return true
	default:
		return false
	}
}
