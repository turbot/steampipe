import HierarchyNode from "./HierarchyNode";
import { CheckNode } from "../index";

class BenchmarkNode extends HierarchyNode {
  constructor(
    sort: string,
    name: string,
    title: string | undefined,
    children?: CheckNode[]
  ) {
    super("benchmark", name, title || name, sort, children || []);
  }
}

export default BenchmarkNode;
