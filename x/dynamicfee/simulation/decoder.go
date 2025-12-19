package simulation

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/kv"
)

// NewDecodeStore returns a decoder function closure that unmarshals the KVPair's
// Value to the corresponding dynamicfee type.
func NewDecodeStore(_ codec.Codec) func(kvA, kvB kv.Pair) string {
	return func(kvA, _ kv.Pair) string {
		panic(fmt.Sprintf("invalid dynamicfee key prefix %X", kvA.Key[:1]))
	}
}
