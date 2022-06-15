import ErrorPanel from "../../Error";
import Inputs, { InputProperties } from "../index";
import { PanelDefinition } from "../../../../hooks/useDashboard";

export type InputDefinition = PanelDefinition & {
  properties: InputProperties;
};

const renderInput = (definition: InputDefinition) => {
  const {
    properties: { unqualified_name: name, type = "select" },
  } = definition;
  const input = Inputs[type];

  if (!input) {
    return <ErrorPanel error={`Unknown input type ${type}`} />;
  }

  const Component = input.component;
  return <Component {...definition} name={name} />;
};

const RenderInput = (props: InputDefinition) => {
  return renderInput(props);
};

export { RenderInput };
