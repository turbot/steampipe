import ErrorPanel from "../../Error";
import Inputs, { InputProperties } from "../index";
import { PanelDefinition } from "../../../../hooks/useReport";

export type InputDefinition = PanelDefinition & {
  properties: InputProperties;
};

const renderInput = (definition: InputDefinition) => {
  const {
    properties: { type },
  } = definition;
  const input = Inputs[type];

  if (!input) {
    return <ErrorPanel error={`Unknown input type ${type}`} />;
  }

  const Component = input.component;
  return <Component {...definition} />;
};

const RenderInput = (props: InputDefinition) => {
  return renderInput(props);
};

export { RenderInput };
