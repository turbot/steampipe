import SelectInput from "../SelectInput";
import { IInput, InputProps } from "../index";

const MultiSelectInput = (props: InputProps) => {
  return <SelectInput {...props} multi />;
};

const definition: IInput = {
  type: "multi",
  component: MultiSelectInput,
};

export default definition;
