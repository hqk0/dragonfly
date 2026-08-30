// Package legacyprotocol implements the Minecraft 1.26.40-1.26.43 wire
// protocol. These releases share protocol ID 2168 with 1.26.44, but differ in
// a small number of packet fields.
package legacyprotocol

import (
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

const (
	protocolID = 2168
	version    = "1.26.40"
)

// Protocol implements minecraft.Protocol for Minecraft 1.26.40-1.26.43.
// Mojang did not change the protocol ID when releasing 1.26.44, so the
// connection's game version is used by gophertunnel to select this protocol.
type Protocol struct{}

var _ minecraft.Protocol = Protocol{}

func (Protocol) ID() int32   { return protocolID }
func (Protocol) Ver() string { return version }
func (Protocol) NewReader(r minecraft.ByteReader, shieldID int32, enableLimits bool) protocol.IO {
	return protocol.NewReader(r, shieldID, enableLimits)
}
func (Protocol) NewWriter(w minecraft.ByteWriter, shieldID int32) protocol.IO {
	return protocol.NewWriter(w, shieldID)
}

func (Protocol) Packets(listener bool) packet.Pool {
	if listener {
		return packet.NewClientPool()
	}
	return packet.NewServerPool()
}

func (Protocol) ConvertToLatest(pk packet.Packet, _ *minecraft.Conn) []packet.Packet {
	switch pk := pk.(type) {
	case *packet.PlayerSkin:
		converted := *pk
		converted.Skin = convertSkin(pk.Skin, pieceTypeToLatest)
		return []packet.Packet{&converted}
	case *packet.PlayerList:
		return []packet.Packet{convertPlayerList(pk, pieceTypeToLatest)}
	default:
		return []packet.Packet{pk}
	}
}

func (Protocol) ConvertFromLatest(pk packet.Packet, _ *minecraft.Conn) []packet.Packet {
	switch pk := pk.(type) {
	case *packet.PlayerSkin:
		converted := *pk
		converted.Skin = convertSkin(pk.Skin, pieceTypeFromLatest)
		return []packet.Packet{&converted}
	case *packet.PlayerList:
		return []packet.Packet{convertPlayerList(pk, pieceTypeFromLatest)}
	default:
		return []packet.Packet{pk}
	}
}

func convertPlayerList(pk *packet.PlayerList, convert func(uint32) uint32) *packet.PlayerList {
	converted := *pk
	converted.Entries = append([]protocol.PlayerListEntry(nil), pk.Entries...)
	for i := range converted.Entries {
		converted.Entries[i].Skin = convertSkin(pk.Entries[i].Skin, convert)
	}
	return &converted
}

func convertSkin(skin protocol.Skin, convert func(uint32) uint32) protocol.Skin {
	converted := skin
	converted.PersonaPieces = append([]protocol.PersonaPiece(nil), skin.PersonaPieces...)
	for i := range converted.PersonaPieces {
		converted.PersonaPieces[i].PieceType = convert(converted.PersonaPieces[i].PieceType)
	}
	return converted
}

// The 1.26.44 protocol inserted Unknown before the existing persona piece
// values and Unsupported after them. Values outside the known ranges are kept
// intact so a future or custom value is not silently destroyed.
func pieceTypeToLatest(pieceType uint32) uint32 {
	if pieceType <= 26 {
		return pieceType + 1
	}
	return pieceType
}

func pieceTypeFromLatest(pieceType uint32) uint32 {
	if pieceType >= protocol.PieceTypeSkeleton && pieceType <= protocol.PieceTypeEmote {
		return pieceType - 1
	}
	return pieceType
}
