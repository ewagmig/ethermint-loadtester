package types

type AccountInfo struct {
	Name     string `json:"name"`
	Address  string `json:"address"`
	Mnemonic string `json:"mnemonic"`
}

type AccountInfosWrapper struct {
	AccountInfos []AccountInfo `json:"account_infos"`
}
