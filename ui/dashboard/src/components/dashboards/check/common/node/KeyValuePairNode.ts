import HierarchyNode from "./HierarchyNode";
import { CheckNodeType, CheckNode } from "../index";

class KeyValuePairNode extends HierarchyNode {
  constructor(
    type: CheckNodeType,
    key: string,
    value: string,
    children?: CheckNode[]
  ) {
    super(type, `${key}=${value}`, value, value, children || []);
  }
}

export default KeyValuePairNode;
