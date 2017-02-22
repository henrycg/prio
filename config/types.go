package config

import (
	"github.com/henrycg/prio/circuit"
)

type FieldType byte

// Maximum sizes for count-min sketch structure.
const MAX_BUCKETS = 1024 * 128
const MAX_HASHES = 256

// Data types that Prio supports.
const (
	TypeInt      FieldType = iota
	TypeIntPow             = iota
	TypeBoolOr             = iota
	TypeBoolAnd            = iota
	TypeCountMin           = iota
	TypeLinReg             = iota

	// Used only for performance comparison. There is
	// no need to use these types in practice.
	TypeIntUnsafe = iota
)

// Definition of a data type that this Prio deployment collects.
type Field struct {
	Name string
	Type FieldType

	IntBits int
	IntPow  int

	CountMinHashes  int
	CountMinBuckets int

	// Each client training example has the form:
	//    (y, x_1, x_2, ..., x_n)
	// where the data has dimension n.
	//
	// The 0th entry is the number of bits in the y value.
	// The rest of the entries represent the number of bits in each x_i.
	LinRegBits []int
}

type ServerAddress struct {
	Public  string `json:"addrPub"`
	Private string `json:"addrPriv"`
}

type ClientAddress string

type Config struct {
	MaxPendingReqs int
	Circuit        *circuit.Circuit
	Fields         []Field
	Servers        []ServerAddress
	Clients        []ClientAddress
}
