import ComboInput from "../ComboInput";
import { IInput, InputProps } from "../types";
import { registerInputComponent } from "../index";

const MultiComboInput = (props: InputProps) => {
  return <ComboInput {...props} multi />;
};

const definition: IInput = {
  type: "multicombo",
  component: MultiComboInput,
};

registerInputComponent(definition.type, definition);

export default definition;
