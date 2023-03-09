import SelectInput from "../SelectInput";
import { IInput, InputProps } from "../types";
import { registerInputComponent } from "../index";

const MultiSelectInput = (props: InputProps) => {
  return <SelectInput {...props} multi />;
};

const definition: IInput = {
  type: "multiselect",
  component: MultiSelectInput,
};

registerInputComponent(definition.type, definition);

export default definition;
