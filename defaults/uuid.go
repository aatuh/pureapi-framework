package defaults

import (
	"sync/atomic"

	"github.com/pureapi/pureapi-util/uuid"
)

// UUID is a string alias that represents a UUID.
type UUID string

// String returns the string representation of the UUID.
//
// Returns:
//   - string: The string representation of the UUID.
func (u UUID) String() string {
	return string(u)
}

// UUIDGen implements the default UUID generation.
type UUIDGen struct{}

// uuidGenFactory holds the UUID environment variable config.
var uuidGenFactory atomic.Value

// init sets the default UUID generator factory.
func init() {
	uuidGenFactory.Store(func() *UUIDGen {
		return &UUIDGen{}
	})
}

// NewUUIDGen returns a new UUID generator using the current factory.
//
// Returns:
//   - *UUIDGen: A new UUID generator.
func NewUUIDGen() *UUIDGen {
	factory, ok := uuidGenFactory.Load().(func() *UUIDGen)
	if !ok || factory == nil {
		return &UUIDGen{}
	}
	return factory()
}

// SetUUIDGenFactory allows overriding the default UUID generator.
//
// Parameters:
//   - factory: The UUID generator factory.
func SetUUIDGenFactory(factory func() *UUIDGen) {
	if factory != nil {
		uuidGenFactory.Store(factory)
	}
}

// Random returns a random UUID or an error.
//
// Returns:
//   - UUID: A random UUID conforming to Version 4 and Variant 1.
//   - error: An error if crypto/rand fails.
func (g *UUIDGen) Random() (UUID, error) {
	val, err := uuid.Ver4Var1()
	return UUID(val.String()), err
}

// MustRandom returns a random UUID and panics on error.
//
// Returns:
//   - UUID: A random UUID conforming to Version 4 and Variant 1.
func (g *UUIDGen) MustRandom() UUID {
	return UUID(uuid.MustVer4Var1().String())
}

// FromString creates a UUID from the given string.
//
// Returns:
//   - UUID: A UUID conforming to Version 4 and Variant 1.
//   - error: An error if the input string is invalid.
func (g *UUIDGen) FromString(s string) (UUID, error) {
	val, err := uuid.Ver4Var1FromString(s)
	return UUID(val.String()), err
}

// MustFromString creates a UUID from the given string, panicking on error.
//
// Parameters:
//   - s: The string to convert to a UUID.
//
// Returns:
//   - UUID: A UUID conforming to Version 4 and Variant 1.
func (g *UUIDGen) MustFromString(s string) UUID {
	return UUID(uuid.MustVer4Var1FromString(s).String())
}

// Zero returns a pointer to a zero-value UUID.
//
// Returns:
//   - *UUID: A pointer to a zero-value UUID.
func (g *UUIDGen) Zero() UUID {
	return UUID(uuid.Zero().String())
}

// IsValid returns true if the given string is a valid UUID.
//
// Parameters:
//   - s: The string to validate.
//
// Returns:
//   - bool: True if the string is a valid UUID, false otherwise.
func (g *UUIDGen) IsValid(s string) bool {
	return uuid.IsValid(s)
}
