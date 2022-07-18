import Graph from "../Graph";
import { GraphProps, IGraph } from "../types";
import { registerGraphComponent } from "../index";

const Sankey = (props: GraphProps) => <Graph {...props} />;

const definition: IGraph = {
  type: "graph",
  component: Sankey,
};

registerGraphComponent(definition.type, definition);

export default definition;
