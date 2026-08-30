package legacyprotocol

import (
	"reflect"
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func TestProtocolMetadataAndPools(t *testing.T) {
	p := Protocol{}
	if got := p.ID(); got != protocolID {
		t.Fatalf("ID() = %d, want %d", got, protocolID)
	}
	if got := p.Ver(); got != version {
		t.Fatalf("Ver() = %q, want %q", got, version)
	}
	if _, ok := p.Packets(true)[packet.IDPlayerSkin]; !ok {
		t.Fatal("listener packet pool does not contain PlayerSkin")
	}
	if _, ok := p.Packets(false)[packet.IDSetScore]().(*packet.SetScore); !ok {
		t.Fatal("dialer packet pool does not use the current SetScore representation")
	}
}

func TestPlayerSkinPieceTypesConvertBothDirections(t *testing.T) {
	latest := &packet.PlayerSkin{Skin: protocol.Skin{PersonaPieces: []protocol.PersonaPiece{
		{PieceType: protocol.PieceTypeSkeleton},
		{PieceType: protocol.PieceTypeEmote},
		{PieceType: protocol.PieceTypeUnknown},
		{PieceType: protocol.PieceTypeUnsupported},
	}}}

	legacy := Protocol{}.ConvertFromLatest(latest, nil)[0].(*packet.PlayerSkin)
	wantLegacy := []uint32{0, 26, protocol.PieceTypeUnknown, protocol.PieceTypeUnsupported}
	assertPieceTypes(t, legacy.Skin.PersonaPieces, wantLegacy)
	assertPieceTypes(t, latest.Skin.PersonaPieces, []uint32{
		protocol.PieceTypeSkeleton,
		protocol.PieceTypeEmote,
		protocol.PieceTypeUnknown,
		protocol.PieceTypeUnsupported,
	})

	oldWire := &packet.PlayerSkin{Skin: protocol.Skin{PersonaPieces: []protocol.PersonaPiece{
		{PieceType: 0},
		{PieceType: 26},
		{PieceType: 27},
	}}}
	converted := Protocol{}.ConvertToLatest(oldWire, nil)[0].(*packet.PlayerSkin)
	assertPieceTypes(t, converted.Skin.PersonaPieces, []uint32{
		protocol.PieceTypeSkeleton,
		protocol.PieceTypeEmote,
		27,
	})
	assertPieceTypes(t, oldWire.Skin.PersonaPieces, []uint32{0, 26, 27})
}

func TestKnownPieceTypeRange(t *testing.T) {
	for latest := uint32(protocol.PieceTypeSkeleton); latest <= protocol.PieceTypeEmote; latest++ {
		legacy := latest - 1
		if got := pieceTypeFromLatest(latest); got != legacy {
			t.Fatalf("pieceTypeFromLatest(%d) = %d, want %d", latest, got, legacy)
		}
		if got := pieceTypeToLatest(legacy); got != latest {
			t.Fatalf("pieceTypeToLatest(%d) = %d, want %d", legacy, got, latest)
		}
	}
	if got := pieceTypeFromLatest(protocol.PieceTypeUnsupported); got != protocol.PieceTypeUnsupported {
		t.Fatalf("unsupported piece type changed to %d", got)
	}
	if got := pieceTypeToLatest(27); got != 27 {
		t.Fatalf("unknown legacy piece type changed to %d", got)
	}
}

func TestPlayerListPieceTypesConvertWithoutMutation(t *testing.T) {
	latest := &packet.PlayerList{Entries: []protocol.PlayerListEntry{{
		Skin: protocol.Skin{PersonaPieces: []protocol.PersonaPiece{{PieceType: protocol.PieceTypeBody}}},
	}}}
	converted := Protocol{}.ConvertFromLatest(latest, nil)[0].(*packet.PlayerList)

	assertPieceTypes(t, converted.Entries[0].Skin.PersonaPieces, []uint32{1})
	assertPieceTypes(t, latest.Entries[0].Skin.PersonaPieces, []uint32{protocol.PieceTypeBody})
}

func TestUnchangedPacketPassesThrough(t *testing.T) {
	original := &packet.Text{}
	if got := (Protocol{}).ConvertToLatest(original, nil)[0]; got != original {
		t.Fatalf("ConvertToLatest returned %p, want original %p", got, original)
	}
	if got := (Protocol{}).ConvertFromLatest(original, nil)[0]; got != original {
		t.Fatalf("ConvertFromLatest returned %p, want original %p", got, original)
	}
}

func assertPieceTypes(t *testing.T, pieces []protocol.PersonaPiece, want []uint32) {
	t.Helper()
	got := make([]uint32, len(pieces))
	for i, piece := range pieces {
		got[i] = piece.PieceType
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("piece types = %v, want %v", got, want)
	}
}
