import SelectInput from "../SelectInput";
import { IInput, InputProps } from "../types";
import { registerInputComponent } from "../index";

const SingleSelectInput = (props: InputProps) => {
  return <SelectInput {...props} />;
};

const definition: IInput = {
  type: "select",
  component: SingleSelectInput,
};

registerInputComponent(definition.type, definition);

export default definition;
