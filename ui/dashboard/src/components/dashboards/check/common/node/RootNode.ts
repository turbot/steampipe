import HierarchyNode from "./HierarchyNode";
import { CheckNode } from "../index";

class RootNode extends HierarchyNode {
  constructor(children?: CheckNode[]) {
    super("root", "root", "Root", "root", children || []);
  }
}

export default RootNode;
