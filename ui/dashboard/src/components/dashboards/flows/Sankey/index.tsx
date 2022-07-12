import Flow from "../Flow";
import { FlowProps, IFlow } from "../types";
import { registerFlowComponent } from "../index";

const Sankey = (props: FlowProps) => <Flow {...props} />;

const definition: IFlow = {
  type: "sankey",
  component: Sankey,
};

registerFlowComponent(definition.type, definition);

export default definition;
