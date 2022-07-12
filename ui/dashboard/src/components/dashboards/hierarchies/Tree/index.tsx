import Hierarchy from "../Hierarchy";
import { HierarchyProps, IHierarchy } from "../types";
import { registerHierarchyComponent } from "../index";

const Tree = (props: HierarchyProps) => <Hierarchy {...props} />;

const definition: IHierarchy = {
  type: "tree",
  component: Tree,
};

registerHierarchyComponent(definition.type, definition);

export default definition;
