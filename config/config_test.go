package config

import (
	"testing"
)

func TestLoad1(t *testing.T) {
	eg := []byte(`[
    {"name": "This"}
  ]`)

	_, err := Load(eg)
	if err == nil {
		t.Fail()
	}
}

func TestLoad2(t *testing.T) {
	eg := []byte(`{
    "servers": [],
    "fields": [
      {"name": "This", "type": "boolOr"}
    ]}
  `)

	f, err := Load(eg)
	if err != nil {
		t.Fatal("Error unmarshalling: ", err)
	}

	if len(f.Fields) != 1 || f.Fields[0].Name != "This" || f.Fields[0].Type != TypeBoolOr {
		t.Fail()
	}
}

func TestLoad3(t *testing.T) {
	eg := []byte(`{
    "servers": [
      {"addrPub": "123"}
    ],
    "fields": [
      {"name": "This", "type": "int", "intBits": 3}
    ]
  }`)

	f, err := Load(eg)
	if err != nil {
		t.Fatal("Error unmarshalling: ", err)
	}

	if len(f.Fields) != 1 || f.Fields[0].Name != "This" || f.Fields[0].Type != TypeInt || f.Fields[0].IntBits != 3 {
		t.Fail()
	}

	if len(f.Servers) != 1 || f.Servers[0].Public != "123" {
		t.Fail()
	}
}

func TestLoad4(t *testing.T) {
	eg := []byte(`{
    "servers": [],
    "fields": [
    {"name": "This", "type": "countMin"}
  ]}`)

	f, err := Load(eg)
	if err == nil || f != nil {
		t.Fail()
	}
}

func TestLoad5(t *testing.T) {
	eg := []byte(`{
    "servers": [
      {"addrPub": "123"}
    ],
    "fields": [
      {"name": "linReg0", 
        "type": "linReg", 
        "FloatLinRegIntercept": 0.02,
        "LinRegBits": [3, 3, 7]
      }
    ]
  }`)

	cfg, err := Load(eg)
	if err != nil {
		t.Fatal("Error unmarshalling: ", err)
	}

	if len(cfg.Fields) != 1 {
		t.Fail()
	}
	f := cfg.Fields[0]

	if f.Name != "linReg0" || f.Type != TypeLinReg {
		t.Fatal("Wrong type")
	}

	if len(f.LinRegBits) != 3 {
		t.FailNow()
	}

	if f.LinRegBits[0] != 3 ||
		f.LinRegBits[1] != 3 ||
		f.LinRegBits[2] != 7 {
		t.Fail()
	}

	/*
		v2 := new(big.Int)
		v2.SetString("7ffffffffffffffffffffffffffff703", 16)
		if f.LinRegTerms[1].Constant.Cmp(v2) != 0 {
			t.Fatalf("Wrong value: %v, wanted %v", f.LinRegTerms[1].Constant, v2)
		}
	*/
}
