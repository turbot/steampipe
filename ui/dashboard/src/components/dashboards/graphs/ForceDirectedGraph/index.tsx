import Graph from "../Graph";
import { GraphProps, IGraph } from "../types";
import { registerGraphComponent } from "../index";

const ForceDirectedGraph = (props: GraphProps) => <Graph {...props} />;

const definition: IGraph = {
  type: "graph",
  component: ForceDirectedGraph,
};

registerGraphComponent(definition.type, definition);

export default definition;
