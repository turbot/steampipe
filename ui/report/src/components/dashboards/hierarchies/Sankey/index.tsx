import Hierarchy from "../Hierarchy";
import { HierarchyProps, IHierarchy } from "../index";

const Sankey = (props: HierarchyProps) => <Hierarchy {...props} />;

const definition: IHierarchy = {
  type: "sankey",
  component: Sankey,
};

export default definition;
