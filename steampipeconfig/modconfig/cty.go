package modconfig

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/terraform/configs/configschema"
	"github.com/zclconf/go-cty/cty"
	ctyjson "github.com/zclconf/go-cty/cty/json"
)

// convert the item into a cty value.
func getCtyValue(item interface{}, block configschema.Block) (cty.Value, error) {
	// get cty spec
	spec := block.DecoderSpec()
	ty := hcldec.ImpliedType(spec)
	json, err := json.Marshal(item)

	if err != nil {
		return cty.EmptyObjectVal, fmt.Errorf("failed to convert cty value - asJson failed %s", err.Error())
	}
	val, err := ctyjson.Unmarshal(json, ty)
	if err != nil {
		return cty.EmptyObjectVal, fmt.Errorf("failed to convert cty value - unmarshal failed %s", err.Error())
	}
	return val, nil
}
