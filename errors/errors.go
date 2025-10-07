package errors

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"sync"

	"github.com/aatuh/pureapi-core/apierror"
)

// CatalogEntry describes a wire error returned by the framework.
type CatalogEntry struct {
	ID      string
	Status  int
	Message string
}

// ErrorCatalog keeps registered catalog entries keyed by ID.
type ErrorCatalog struct {
	mu      sync.RWMutex
	entries map[string]CatalogEntry
}

// NewErrorCatalog constructs an error catalog with optional initial entries.
func NewErrorCatalog(entries ...CatalogEntry) (*ErrorCatalog, error) {
	catalog := &ErrorCatalog{entries: make(map[string]CatalogEntry, len(entries))}
	for _, entry := range entries {
		if err := catalog.Register(entry); err != nil {
			return nil, err
		}
	}
	return catalog, nil
}

// DefaultErrorCatalog returns a catalog with core framework entries.
func DefaultErrorCatalog() *ErrorCatalog {
	catalog, _ := NewErrorCatalog(
		CatalogEntry{ID: "internal_error", Status: http.StatusInternalServerError, Message: "Internal server error"},
		CatalogEntry{ID: "invalid_request", Status: http.StatusBadRequest, Message: "Request validation failed"},
		CatalogEntry{ID: "unauthorized", Status: http.StatusUnauthorized, Message: "Unauthorized"},
		CatalogEntry{ID: "forbidden", Status: http.StatusForbidden, Message: "Forbidden"},
	)
	return catalog
}

// Register adds an entry to the catalog ensuring uniqueness.
func (c *ErrorCatalog) Register(entry CatalogEntry) error {
	if entry.ID == "" {
		return fmt.Errorf("catalog entry id must not be empty")
	}
	if entry.Status < 400 || entry.Status > 599 {
		return fmt.Errorf("catalog entry %s has invalid HTTP status %d", entry.ID, entry.Status)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, exists := c.entries[entry.ID]; exists {
		return fmt.Errorf("catalog entry with id %s already registered", entry.ID)
	}
	copied := entry
	if copied.Message == "" {
		copied.Message = http.StatusText(entry.Status)
	}
	c.entries[entry.ID] = copied
	return nil
}

// Lookup returns the catalog entry for id.
func (c *ErrorCatalog) Lookup(id string) (CatalogEntry, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, ok := c.entries[id]
	return entry, ok
}

// Clone produces a shallow copy of the catalog.
func (c *ErrorCatalog) Clone() *ErrorCatalog {
	c.mu.RLock()
	defer c.mu.RUnlock()
	copy := &ErrorCatalog{entries: make(map[string]CatalogEntry, len(c.entries))}
	for k, v := range c.entries {
		copy.entries[k] = v
	}
	return copy
}

// wireMessage allows errors to override the catalog message.
type wireMessage interface {
	WireMessage() string
}

// wireData allows errors to attach extra payload.
type wireData interface {
	WireData() any
}

type typeRegistration struct {
	typ     reflect.Type
	entryID string
}

type isRegistration struct {
	target  error
	entryID string
}

// ErrorMapper maps Go errors to catalog entries.
type ErrorMapper struct {
	catalog   *ErrorCatalog
	defaultID string

	mu       sync.RWMutex
	typeRegs []typeRegistration
	isRegs   []isRegistration
}

// NewErrorMapper creates a mapper backed by catalog. defaultID must exist.
func NewErrorMapper(catalog *ErrorCatalog, defaultID string) (*ErrorMapper, error) {
	if catalog == nil {
		return nil, fmt.Errorf("catalog must not be nil")
	}
	if defaultID == "" {
		return nil, fmt.Errorf("default error id must not be empty")
	}
	if _, ok := catalog.Lookup(defaultID); !ok {
		return nil, fmt.Errorf("default error id %s not present in catalog", defaultID)
	}
	return &ErrorMapper{catalog: catalog, defaultID: defaultID}, nil
}

// RegisterType registers errors matched via errors.As against a catalog entry.
func (m *ErrorMapper) RegisterType(prototype any, entryID string) error {
	if prototype == nil {
		return fmt.Errorf("prototype must not be nil")
	}
	if err := m.ensureEntry(entryID); err != nil {
		return err
	}
	typ := reflect.TypeOf(prototype)
	m.mu.Lock()
	defer m.mu.Unlock()
	m.typeRegs = append(m.typeRegs, typeRegistration{typ: typ, entryID: entryID})
	return nil
}

// RegisterIs registers a sentinel error matched using errors.Is.
func (m *ErrorMapper) RegisterIs(target error, entryID string) error {
	if target == nil {
		return fmt.Errorf("target must not be nil")
	}
	if err := m.ensureEntry(entryID); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.isRegs = append(m.isRegs, isRegistration{target: target, entryID: entryID})
	return nil
}

// ensureEntry checks that an entry exists in catalog.
func (m *ErrorMapper) ensureEntry(entryID string) error {
	if entryID == "" {
		return fmt.Errorf("entry id must not be empty")
	}
	if _, ok := m.catalog.Lookup(entryID); !ok {
		return fmt.Errorf("entry id %s not found in catalog", entryID)
	}
	return nil
}

// CatalogError allows custom errors to dictate the catalog id directly.
type CatalogError interface {
	CatalogID() string
}

// MappedError represents a mapped error ready for rendering.
type MappedError struct {
	Entry   CatalogEntry
	Message string
	Data    any
	Cause   error
}

// Map resolves err into a catalog entry. Nil err maps to default.
func (m *ErrorMapper) Map(err error) MappedError {
	if err == nil {
		entry, _ := m.catalog.Lookup(m.defaultID)
		return MappedError{Entry: entry, Message: entry.Message}
	}

	entry := m.matchEntry(err)
	msg := entry.Message
	if wm, ok := err.(wireMessage); ok {
		if custom := wm.WireMessage(); custom != "" {
			msg = custom
		}
	}
	var data any
	if wd, ok := err.(wireData); ok {
		data = wd.WireData()
	}
	return MappedError{Entry: entry, Message: msg, Data: data, Cause: err}
}

func (m *ErrorMapper) matchEntry(err error) CatalogEntry {
	if err == nil {
		entry, _ := m.catalog.Lookup(m.defaultID)
		return entry
	}

	if ce, ok := err.(CatalogError); ok {
		if entry, ok := m.catalog.Lookup(ce.CatalogID()); ok {
			return entry
		}
	}

	if entry, ok := m.matchIs(err); ok {
		return entry
	}
	if entry, ok := m.matchType(err); ok {
		return entry
	}

	entry, _ := m.catalog.Lookup(m.defaultID)
	return entry
}

func (m *ErrorMapper) matchIs(err error) (CatalogEntry, bool) {
	m.mu.RLock()
	regs := append([]isRegistration(nil), m.isRegs...)
	m.mu.RUnlock()
	for _, reg := range regs {
		if errors.Is(err, reg.target) {
			if entry, ok := m.catalog.Lookup(reg.entryID); ok {
				return entry, true
			}
		}
	}
	return CatalogEntry{}, false
}

func (m *ErrorMapper) matchType(err error) (CatalogEntry, bool) {
	m.mu.RLock()
	regs := append([]typeRegistration(nil), m.typeRegs...)
	m.mu.RUnlock()
	for _, reg := range regs {
		if matchesAs(err, reg.typ) {
			if entry, ok := m.catalog.Lookup(reg.entryID); ok {
				return entry, true
			}
		}
	}
	return CatalogEntry{}, false
}

func matchesAs(err error, typ reflect.Type) bool {
	if err == nil {
		return false
	}
	target := reflect.New(typ)
	return errors.As(err, target.Interface())
}

// RenderError prepares the API error payload for the mapped error.
func RenderError(mapped MappedError) *apierror.DefaultAPIError {
	payload := apierror.NewAPIError(mapped.Entry.ID).WithMessage(mapped.Message)
	if mapped.Data != nil {
		payload = payload.WithData(mapped.Data)
	}
	return payload
}
