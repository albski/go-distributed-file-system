package p2p

const (
	MessageRPC = 0x01
	StreamRPC  = 0x02
)

type RPC struct {
	From    string
	Payload []byte
	Stream  bool
}
