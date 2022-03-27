import SelectInput from "../SelectInput";
import { IInput, InputProps } from "../index";

const MultiSelectInput = (props: InputProps) => {
  return <SelectInput {...props} />;
};

const definition: IInput = {
  type: "multiselect",
  component: MultiSelectInput,
};

export default definition;
