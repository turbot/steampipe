import Hierarchy from "../Hierarchy";
import { HierarchyProps, IHierarchy } from "../types";

const Tree = (props: HierarchyProps) => <Hierarchy {...props} />;

const definition: IHierarchy = {
  type: "tree",
  component: Tree,
};

export default definition;
