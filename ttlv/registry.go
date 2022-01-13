package ttlv

import (
	"github.com/ansel1/merry"
	"github.com/gsealy/kmip-go/internal/kmiputil"
	"sort"
)

// DefaultRegistry holds the default mappings of types, tags, enums, and bitmasks
// to canonical names and normalized names from the KMIP spec.  It is pre-populated with the 1.4 spec's
// values.  It can be replaced, or additional values can be registered with it.
//
// It is not currently concurrent-safe, so replace or configure it early in your
// program.
var DefaultRegistry Registry

// nolint:gochecknoinits
func init() {
	RegisterTypes(&DefaultRegistry)
}

var ErrInvalidHexString = kmiputil.ErrInvalidHexString
var ErrUnregisteredEnumName = merry.New("unregistered enum name")

// NormalizeName tranforms KMIP names from the spec into the
// normalized form of the name.  Typically, this means removing spaces,
// and replacing some special characters.  The normalized form of the name
// is used in the JSON and XML encodings from the KMIP Profiles.
// The spec describes the normalization process in 5.4.1.1 and 5.5.1.1
func NormalizeName(s string) string {
	return kmiputil.NormalizeName(s)
}

// Enum represents an enumeration of KMIP values (as uint32), and maps them
// to the canonical string names and the normalized string names of the
// value as declared in the KMIP specs.
// Enum is used to transpose values from strings to byte values, as required
// by the JSON and XML encodings defined in the KMIP Profiles spec.
// These mappings are also used to pretty print KMIP values, and to marshal
// and unmarshal enum and bitmask values to golang string values.
//
// Enum currently uses plain maps, so it is not thread safe to register new values
// concurrently.  You should register all values at the start of your program before
// using this package concurrently.
//
// Enums are used in the KMIP spec for two purposes: for defining the possible values
// for values encoded as the KMIP Enumeration type, and for bitmask values.  Bitmask
// values are encoded as Integers, but are really enum values bitwise-OR'd together.
//
// Enums are registered with a Registry.  The code to register enums is typically
// generated by the kmipgen tool.
type Enum struct {
	valuesToName          map[uint32]string
	valuesToCanonicalName map[uint32]string
	nameToValue           map[string]uint32
	canonicalNamesToValue map[string]uint32
	bitMask               bool
}

func NewEnum() Enum {
	return Enum{}
}

func NewBitmask() Enum {
	return Enum{
		bitMask: true,
	}
}

// RegisterValue adds a mapping of a uint32 value to a name.  The name will be
// processed by NormalizeName to produce the normalized enum value name as described
// in the KMIP spec.
func (e *Enum) RegisterValue(v uint32, name string) {
	nn := NormalizeName(name)
	if e.valuesToName == nil {
		e.valuesToName = map[uint32]string{}
		e.nameToValue = map[string]uint32{}
		e.valuesToCanonicalName = map[uint32]string{}
		e.canonicalNamesToValue = map[string]uint32{}
	}
	e.valuesToName[v] = nn
	e.nameToValue[nn] = v
	e.valuesToCanonicalName[v] = name
	e.canonicalNamesToValue[name] = v
}

func (e *Enum) Name(v uint32) (string, bool) {
	if e == nil {
		return "", false
	}
	name, ok := e.valuesToName[v]
	return name, ok
}

func (e *Enum) CanonicalName(v uint32) (string, bool) {
	if e == nil {
		return "", false
	}
	name, ok := e.valuesToCanonicalName[v]
	return name, ok
}

func (e *Enum) Value(name string) (uint32, bool) {
	if e == nil {
		return 0, false
	}
	v, ok := e.nameToValue[name]
	if !ok {
		v, ok = e.canonicalNamesToValue[name]
	}
	return v, ok
}

func (e *Enum) Values() []uint32 {
	values := make([]uint32, 0, len(e.valuesToName))
	for v := range e.valuesToName {
		values = append(values, v)
	}
	// Always list them in order of value so output is stable.
	sort.Sort(uint32Slice(values))
	return values
}

func (e *Enum) Bitmask() bool {
	if e == nil {
		return false
	}
	return e.bitMask
}

// Registry holds all the known tags, types, enums and bitmaps declared in
// a KMIP spec.  It's used throughout the package to map values their canonical
// and normalized names.
type Registry struct {
	enums map[Tag]EnumMap
	tags  Enum
	types Enum
}

func (r *Registry) RegisterType(t Type, name string) {
	r.types.RegisterValue(uint32(t), name)
}

func (r *Registry) RegisterTag(t Tag, name string) {
	r.tags.RegisterValue(uint32(t), name)
}

func (r *Registry) RegisterEnum(t Tag, def EnumMap) {
	if r.enums == nil {
		r.enums = map[Tag]EnumMap{}
	}
	r.enums[t] = def
}

// EnumForTag returns the enum map registered for a tag.  Returns
// nil if no map is registered for this tag.
func (r *Registry) EnumForTag(t Tag) EnumMap {
	if r.enums == nil {
		return nil
	}
	return r.enums[t]
}

func (r *Registry) IsBitmask(t Tag) bool {
	if e := r.EnumForTag(t); e != nil {
		return e.Bitmask()
	}
	return false
}

func (r *Registry) IsEnum(t Tag) bool {
	if e := r.EnumForTag(t); e != nil {
		return !e.Bitmask()
	}
	return false
}

func (r *Registry) Tags() EnumMap {
	return &r.tags
}

func (r *Registry) Types() EnumMap {
	return &r.types
}

func (r *Registry) FormatEnum(t Tag, v uint32) string {
	return FormatEnum(v, r.EnumForTag(t))
}

func (r *Registry) FormatInt(t Tag, v int32) string {
	return FormatInt(v, r.EnumForTag(t))
}

func (r *Registry) FormatTag(t Tag) string {
	return FormatTag(uint32(t), &r.tags)
}

func (r *Registry) FormatTagCanonical(t Tag) string {
	return FormatTagCanonical(uint32(t), &r.tags)
}

func (r *Registry) FormatType(t Type) string {
	return FormatType(byte(t), &r.types)
}

func (r *Registry) ParseEnum(t Tag, s string) (uint32, error) {
	return ParseEnum(s, r.EnumForTag(t))
}

func (r *Registry) ParseInt(t Tag, s string) (int32, error) {
	return ParseInt(s, r.EnumForTag(t))
}

// returns TagNone if not found.
// returns error if s is a malformed hex string, or a hex string of incorrect length
func (r *Registry) ParseTag(s string) (Tag, error) {
	return ParseTag(s, &r.tags)
}

func (r *Registry) ParseType(s string) (Type, error) {
	return ParseType(s, &r.types)
}

// uint32Slice attaches the methods of Interface to []int, sorting in increasing order.
type uint32Slice []uint32

func (p uint32Slice) Len() int           { return len(p) }
func (p uint32Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p uint32Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
