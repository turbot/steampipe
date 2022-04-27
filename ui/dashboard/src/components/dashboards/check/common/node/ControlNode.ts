import HierarchyNode from "./HierarchyNode";
import { CheckNode } from "../index";

class ControlNode extends HierarchyNode {
  constructor(
    sort: string,
    name: string,
    title: string | undefined,
    children?: CheckNode[]
  ) {
    super("control", name, title || name, sort, children || []);
  }
}

export default ControlNode;
