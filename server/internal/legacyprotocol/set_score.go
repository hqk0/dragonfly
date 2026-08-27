package legacyprotocol

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// setScore is the 1.26.40-1.26.43 representation of packet.SetScore. In
// 1.26.44, the optional objective name of a removed entry became doubly
// optional without receiving a new protocol ID.
type setScore struct {
	Entries []scoreboardEntry
}

func (*setScore) ID() uint32 { return packet.IDSetScore }

func (pk *setScore) Marshal(io protocol.IO) {
	protocol.Slice(io, &pk.Entries)
}

type scoreboardEntry struct {
	EntryID        int64
	ObjectiveName  string
	Score          int32
	IdentityType   byte
	EntityUniqueID int64
	DisplayName    string
}

func (entry *scoreboardEntry) Marshal(io protocol.IO) {
	variant := uint32(entry.IdentityType)
	io.Varuint32(&variant)
	entry.IdentityType = byte(variant)

	typeNames := [...]string{"remove", "changeplayer", "changeentity", "changefakeplayer"}
	if variant >= uint32(len(typeNames)) {
		io.UnknownEnumOption(variant, "scoreboard entry variant")
		return
	}
	typeName := typeNames[variant]
	io.String(&typeName)
	io.Varint64(&entry.EntryID)

	switch entry.IdentityType {
	case protocol.ScoreboardIdentityRemove:
		objective := protocol.Optional[string]{}
		if entry.ObjectiveName != "" {
			objective = protocol.Option(entry.ObjectiveName)
		}
		protocol.OptionalFunc(io, &objective, io.String)
		entry.ObjectiveName, _ = objective.Value()
	case protocol.ScoreboardIdentityEntity, protocol.ScoreboardIdentityPlayer:
		io.String(&entry.ObjectiveName)
		io.Int32(&entry.Score)
		io.ActorUniqueID(&entry.EntityUniqueID)
	case protocol.ScoreboardIdentityFakePlayer:
		io.String(&entry.ObjectiveName)
		io.Int32(&entry.Score)
		io.String(&entry.DisplayName)
	}
}

func setScoreFromLatest(pk *packet.SetScore) *setScore {
	converted := &setScore{Entries: make([]scoreboardEntry, len(pk.Entries))}
	for i, entry := range pk.Entries {
		converted.Entries[i] = scoreboardEntry{
			EntryID:        entry.EntryID,
			ObjectiveName:  entry.ObjectiveName,
			Score:          entry.Score,
			IdentityType:   entry.IdentityType,
			EntityUniqueID: entry.EntityUniqueID,
			DisplayName:    entry.DisplayName,
		}
	}
	return converted
}

func setScoreToLatest(pk *setScore) *packet.SetScore {
	converted := &packet.SetScore{Entries: make([]protocol.ScoreboardEntry, len(pk.Entries))}
	for i, entry := range pk.Entries {
		converted.Entries[i] = protocol.ScoreboardEntry{
			EntryID:        entry.EntryID,
			ObjectiveName:  entry.ObjectiveName,
			Score:          entry.Score,
			IdentityType:   entry.IdentityType,
			EntityUniqueID: entry.EntityUniqueID,
			DisplayName:    entry.DisplayName,
		}
	}
	return converted
}
