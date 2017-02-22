// Package config represents the global configuration
// options for a Prio deployment.
package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
)

// Each server buffers a certain number of client submissions.
// This value defines how large that buffer is by default.
const DEFAULT_MAX_PENDING_REQS = 64

// The default configuration.
var Default *Config

// See: https://talks.golang.org/2015/json.slide#23
func (ft *FieldType) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	got, ok := map[string]FieldType{
		"int":       TypeInt,
		"intPow":    TypeIntPow,
		"intUnsafe": TypeIntUnsafe,
		"boolOr":    TypeBoolOr,
		"boolAnd":   TypeBoolAnd,
		"countMin":  TypeCountMin,
		"linReg":    TypeLinReg,
	}[s]
	if !ok {
		return fmt.Errorf("Invalid FieldType %q", s)
	}
	*ft = got
	return nil
}

func checkInt(f *Field) error {
	if f.IntBits <= 0 {
		return fmt.Errorf("Field of type int or intPow must have intBits > 0")
	}

	if f.IntBits > 64 {
		return fmt.Errorf("We only support up to 64-bit ints")
	}

	return nil
}

func checkIntPow(f *Field) error {
	valid := f.IntPow == 2 || f.IntPow == 4 || f.IntPow == 8

	err := checkInt(f)
	if err != nil {
		return err
	}

	if !valid {
		fmt.Errorf("Int power must be one of {2, 4, 8}")
	}

	return nil
}

func checkCountMin(f *Field) error {
	if f.CountMinHashes <= 0 {
		return fmt.Errorf("Field of type countMin must have countMinHashes > 0")
	}

	if f.CountMinBuckets <= 0 {
		return fmt.Errorf("Field of type countMin must have countMinBuckets > 0")
	}

	if f.CountMinHashes >= MAX_HASHES {
		return fmt.Errorf("Field of type countMin must have countMinBuckets < %v", MAX_HASHES)
	}

	if f.CountMinBuckets >= MAX_BUCKETS {
		return fmt.Errorf("Field of type countMin must have countMinBuckets < %v", MAX_BUCKETS)
	}

	return nil
}

func checkLinReg(f *Field) error {
	if len(f.LinRegBits) < 2 {
		return fmt.Errorf("LinReg must have at least two terms")
	}

	for t := 0; t < len(f.LinRegBits); t++ {
		if len(f.LinRegBits) <= 0 {
			log.Fatal("LinRegBits[%v] must be at least 1", t)
		}
	}

	return nil
}

func LoadFile(filename string) *Config {
	if len(filename) == 0 {
		return Default
	}

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("Error reading file: %v", err)
		return nil
	}

	cfg, err := Load(data)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
		return nil
	}

	return cfg
}

// Load a configuration file from a byte array.
func Load(s []byte) (*Config, error) {
	cfg := new(Config)

	// XXX Here for now
	//cfg.Fields = make([]Field, 0)
	err := json.Unmarshal(s, &cfg)
	if err != nil {
		return nil, err
	}

	if cfg.MaxPendingReqs == 0 {
		cfg.MaxPendingReqs = DEFAULT_MAX_PENDING_REQS
	}

	for i := 0; i < len(cfg.Fields); i++ {
		f := &cfg.Fields[i]
		switch f.Type {
		case TypeInt:
			err = checkInt(f)
		case TypeIntPow:
			err = checkIntPow(f)
		case TypeIntUnsafe:
			continue
		case TypeCountMin:
			err = checkCountMin(f)
		case TypeLinReg:
			err = checkLinReg(f)
		}
		if err != nil {
			return nil, err
		}
	}
	return cfg, err
}

func (cfg *Config) NumServers() int {
	return len(cfg.Servers)
}

func init() {
	var err error
	//Default, err = Load([]byte(config_Int(5, 128)))
	Default, err = Load([]byte(config_Mixed()))

	if err != nil {
		log.Fatalf("error: %v", err)
	}
}
