package config

import (
	"fmt"
)

func config_Mixed() string {
	return `{
    "servers": [
      {"addrPub": "localhost:9000", "addrPriv": "localhost:9050"},
      {"addrPub": "localhost:9001", "addrPriv": "localhost:9051"},
      {"addrPub": "localhost:9002", "addrPriv": "localhost:9052"},
      {"addrPub": "localhost:9003", "addrPriv": "localhost:9053"},
      {"addrPub": "localhost:9004", "addrPriv": "localhost:9054"}
    ],
    "fields": [
      {"name": "val0", "type": "int", "intBits": 4},
      {"name": "bool0", "type": "boolOr"},
      {"name": "bool1", "type": "boolAnd"},
      {"name": "unsafe0", "type": "intUnsafe", "intBits": 5},
      {"name": "pow0", "type": "intPow", "intPow": 4, "intBits": 3},
			{"name": "sketch", "type": "countMin",
	       "countMinBuckets": 32,
	       "countMinHashes": 8},
			{"name": "linReg0", "type": "linReg",
	        "linRegBits": [2,3,4,5,6]}
    ]
  }`
}

func config_IntUnsafe(nBits int, nRepeat int) string {
	return repeat(fmt.Sprintf(`"type": "intUnsafe", "intBits": %v`, nBits), nRepeat)
}

func config_Int(nBits int, nRepeat int) string {
	return repeat(fmt.Sprintf(`"type": "int", "intBits": %v`, nBits), nRepeat)
}

func repeat(t string, how_many int) string {
	out := `{
    "servers": [
      {"addrPub": "localhost:9000", "addrPriv": "localhost:9050"},
      {"addrPub": "localhost:9001", "addrPriv": "localhost:9051"},
      {"addrPub": "localhost:9002", "addrPriv": "localhost:9052"},
      {"addrPub": "localhost:9003", "addrPriv": "localhost:9053"},
      {"addrPub": "localhost:9004", "addrPriv": "localhost:9054"}
    ],
    "fields": [
  `
	for i := 0; i < how_many; i++ {
		out = out + fmt.Sprintf(`{"name": "val%v", %v}`, i, t)
		if i != how_many-1 {
			out = out + ","
		}
	}
	out = out + "]}"
	return out
}
