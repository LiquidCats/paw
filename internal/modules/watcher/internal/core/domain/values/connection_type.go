package values


import "slices"

type ConnectionType string

const (
	ConnectionTypeRPC ConnectionType = "rpc"
	ConnectionTypeREST ConnectionType = "rest"
	ConnectionTypeWS  ConnectionType = "ws"
	ConnectionTypeP2P ConnectionType = "p2p"
)

func (c ConnectionType) String() string {
	return string(c)
}

func (c ConnectionType) Equals(other ConnectionType) bool {
	return c == other
}

func (c ConnectionType) OneOf(other ...ConnectionType) bool {
	return slices.ContainsFunc(other, c.Equals)
}

func (c ConnectionType) IsValid() bool {
	switch c {
	case ConnectionTypeRPC,
		ConnectionTypeREST,
		ConnectionTypeWS,
		ConnectionTypeP2P:
		return true
	default:
		return false
	}
}
