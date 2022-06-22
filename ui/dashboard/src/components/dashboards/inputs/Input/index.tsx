import ErrorPanel from "../../Error";
import Inputs, { InputProperties } from "../index";
import { PanelDefinition } from "../../../../types/panel";

export type InputDefinition = PanelDefinition & {
  properties: InputProperties;
};

const renderInput = (definition: InputDefinition) => {
  const {
    display_type = "select",
    properties: { unqualified_name: name },
  } = definition;
  const input = Inputs[display_type];

  if (!input) {
    return <ErrorPanel error={`Unknown input type ${display_type}`} />;
  }

  const Component = input.component;
  return <Component {...definition} name={name} />;
};

const RenderInput = (props: InputDefinition) => {
  return renderInput(props);
};

export { RenderInput };
