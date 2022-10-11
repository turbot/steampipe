import ErrorPanel from "../../Error";
import { getInputComponent } from "../index";
import { InputProperties } from "../types";
import { PanelDefinition } from "../../../../types";
import { registerComponent } from "../../index";

export type InputDefinition = PanelDefinition & {
  properties: InputProperties;
};

const renderInput = (definition: InputDefinition) => {
  const {
    display_type = "select",
    properties: { unqualified_name: name },
  } = definition;
  const input = getInputComponent(display_type);

  if (!input) {
    return <ErrorPanel error={`Unknown input type ${display_type}`} />;
  }

  const Component = input.component;
  return <Component {...definition} name={name} />;
};

const RenderInput = (props: InputDefinition) => {
  return renderInput(props);
};

registerComponent("input", RenderInput);
