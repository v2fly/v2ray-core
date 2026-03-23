package rrpitTransferLane

const ReconstructionLengthFieldSize = reconstructionLengthFieldSize

type ReconstructionData struct {
	LengthOfData uint16 `struc:"uint16,sizeof=Data"`
	Data         []byte
	// For the data feed into raptorQ FEC, this data is always padded to shard size
}
type TransferData struct {
	// This value shows the number of data shards in this lane,
	// since we don't delay the transfer of payload, this value is 0 for data packet
	TotalDataShards uint32
	// this is the sequence of this packet, first TotalDataShards packets are data packet
	// rest of them are reconstruction packet. The value of TotalDataShards is not present in every packet.
	Seq          uint32
	LengthOfData uint16 `struc:"uint16,sizeof=Data"`
	Data         []byte
}

type TransferControl struct {
	// tell sender how many chunk the receiver have seen, includes both data and reconstruction shards
	SeenChunks uint16
}
