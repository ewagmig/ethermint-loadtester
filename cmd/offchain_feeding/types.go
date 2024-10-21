package offchain_feeding

type BaseAccount struct {
	Address       string      `json:"address"`
	PubKey        interface{} `json:"pub_key"`
	AccountNumber string      `json:"account_number"`
	Sequence      string      `json:"sequence"`
}

type EthAccount struct {
	Type        string      `json:"@type"`
	BaseAccount BaseAccount `json:"base_account"`
	CodeHash    string      `json:"code_hash"`
}

type Coin struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

type Balance struct {
	Address string `json:"address"`
	Coins   []Coin `json:"coins"`
}
