### Hi there! 👋

This project aims to make a peer-to-peer file system in Go capable of handling and streaming large files efficiently.

### Explanation
#### `p2p` module
The function of this module is simple and quite generic, abstractions can be found in the `transport.go` file. These abstractions were created, not discovered. The RPC is based on the peek buffer that should be sent before actual encoded data; 0x01 if it's a message, 0x02 in case of stream.
