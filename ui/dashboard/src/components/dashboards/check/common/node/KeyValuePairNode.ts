import HierarchyNode from "./HierarchyNode";
import { CheckNodeType, CheckNode } from "../index";

class KeyValuePairNode extends HierarchyNode {
  constructor(
    sort: string,
    type: CheckNodeType,
    key: string,
    value: string,
    children?: CheckNode[]
  ) {
    super(type, `${key}=${value}`, value, sort, children || []);
  }
}

export default KeyValuePairNode;
