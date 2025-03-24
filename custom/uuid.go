package custom

import (
	"github.com/pureapi/pureapi-util/uuid"
)

type UUID = uuid.UUID

type UUIDGen struct{}

func NewUUIDGen() *UUIDGen {
	return &UUIDGen{}
}

func (g *UUIDGen) Random() (UUID, error) {
	return uuid.Ver4Var1()
}

func (g *UUIDGen) MustRandom() UUID {
	return uuid.MustVer4Var1()
}

func (g *UUIDGen) FromString(s string) (UUID, error) {
	return uuid.Ver4Var1FromString(s)
}

func (g *UUIDGen) MustFromString(s string) *UUID {
	return uuid.MustVer4Var1FromString(s)
}

func (g *UUIDGen) Zero() *UUID {
	return uuid.Zero()
}

func (g *UUIDGen) IsValid(s string) bool {
	return uuid.IsValid(s)
}
