import SelectInput from "../SelectInput";
import { IInput, InputProps } from "../index";

const SingleSelectInput = (props: InputProps) => {
  return <SelectInput {...props} />;
};

const definition: IInput = {
  type: "select",
  component: SingleSelectInput,
};

export default definition;
