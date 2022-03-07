import Flow from "../Flow";
import { FlowProps, IFlow } from "../index";

const Sankey = (props: FlowProps) => <Flow {...props} />;

const definition: IFlow = {
  type: "sankey",
  component: Sankey,
};

export default definition;
