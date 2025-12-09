package core

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/wormhole-foundation/wormhole-go-sdk/types"
)

// VAA represents a Verified Action Approval
type VAA struct {
	Version          uint8
	GuardianSetIndex uint32
	Signatures       []Signature
	Timestamp        time.Time
	Nonce            uint32
	EmitterChain     types.ChainID
	EmitterAddress   types.UniversalAddress
	Sequence         uint64
	ConsistencyLevel uint8
	Payload          []byte
}

// Signature represents a guardian signature
type Signature struct {
	Index     uint8
	Signature [65]byte
}

// ParseVAA parses a VAA from bytes
func ParseVAA(data []byte) (*VAA, error) {
	if len(data) < 6 {
		return nil, fmt.Errorf("VAA too short")
	}

	vaa := &VAA{
		Version:          data[0],
		GuardianSetIndex: binary.BigEndian.Uint32(data[1:5]),
	}

	sigLen := int(data[5])
	offset := 6

	// Parse signatures
	vaa.Signatures = make([]Signature, sigLen)
	for i := 0; i < sigLen; i++ {
		if offset+66 > len(data) {
			return nil, fmt.Errorf("invalid signature data")
		}
		sig := Signature{
			Index: data[offset],
		}
		copy(sig.Signature[:], data[offset+1:offset+66])
		vaa.Signatures[i] = sig
		offset += 66
	}

	// Parse body
	if offset+51 > len(data) {
		return nil, fmt.Errorf("invalid VAA body")
	}

	vaa.Timestamp = time.Unix(int64(binary.BigEndian.Uint32(data[offset:offset+4])), 0)
	vaa.Nonce = binary.BigEndian.Uint32(data[offset+4 : offset+8])
	vaa.EmitterChain = types.ChainID(binary.BigEndian.Uint16(data[offset+8 : offset+10]))

	var emitterAddr types.UniversalAddress
	copy(emitterAddr[:], data[offset+10:offset+42])
	vaa.EmitterAddress = emitterAddr

	vaa.Sequence = binary.BigEndian.Uint64(data[offset+42 : offset+50])
	vaa.ConsistencyLevel = data[offset+50]
	vaa.Payload = data[offset+51:]

	return vaa, nil
}

// Serialize serializes the VAA to bytes
func (v *VAA) Serialize() ([]byte, error) {
	// Calculate total size
	size := 6 + len(v.Signatures)*66 + 51 + len(v.Payload)
	data := make([]byte, size)

	// Header
	data[0] = v.Version
	binary.BigEndian.PutUint32(data[1:5], v.GuardianSetIndex)
	data[5] = uint8(len(v.Signatures))

	offset := 6
	// Signatures
	for _, sig := range v.Signatures {
		data[offset] = sig.Index
		copy(data[offset+1:offset+66], sig.Signature[:])
		offset += 66
	}

	// Body
	binary.BigEndian.PutUint32(data[offset:offset+4], uint32(v.Timestamp.Unix()))
	binary.BigEndian.PutUint32(data[offset+4:offset+8], v.Nonce)
	binary.BigEndian.PutUint16(data[offset+8:offset+10], uint16(v.EmitterChain))
	copy(data[offset+10:offset+42], v.EmitterAddress[:])
	binary.BigEndian.PutUint64(data[offset+42:offset+50], v.Sequence)
	data[offset+50] = v.ConsistencyLevel
	copy(data[offset+51:], v.Payload)

	return data, nil
}

// Hash returns the VAA hash
func (v *VAA) Hash() []byte {
	// In a real implementation, this would compute the Keccak256 hash
	return nil
}
