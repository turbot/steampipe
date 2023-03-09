import ComboInput from "../ComboInput";
import { IInput, InputProps } from "../types";
import { registerInputComponent } from "../index";

const SingleComboInput = (props: InputProps) => {
  return <ComboInput {...props} />;
};

const definition: IInput = {
  type: "combo",
  component: SingleComboInput,
};

registerInputComponent(definition.type, definition);

export default definition;
