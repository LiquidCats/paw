package entities

type Network string

type NetParams struct {
	Bech32HRPSegwit         string `yaml:"Bech32HRPSegwit"`
	Bech32HRPMweb           string `yaml:"Bech32HRPMweb"`
	PubKeyHashAddrID        byte   `yaml:"PubKeyHashAddrID"`
	ScriptHashAddrID        byte   `yaml:"ScriptHashAddrID"`
	WitnessPubKeyHashAddrID byte   `yaml:"WitnessPubKeyHashAddrID"`
	WitnessScriptHashAddrID byte   `yaml:"WitnessScriptHashAddrID"`
}
