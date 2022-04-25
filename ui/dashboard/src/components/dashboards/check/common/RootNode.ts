import {
  CheckNodeStatus,
  CheckNodeType,
  CheckSummary,
  CheckNode,
} from "./index";

class RootNode implements CheckNode {
  private readonly _children: CheckNode[];

  constructor(children?: CheckNode[]) {
    this._children = children || [];
  }

  get depth(): number {
    return 0;
  }

  get name(): string {
    return "root";
  }

  get title(): string {
    return "Root";
  }

  get type(): CheckNodeType {
    return "root";
  }

  get summary(): CheckSummary {
    const summary = {
      alarm: 0,
      ok: 0,
      info: 0,
      skip: 0,
      error: 0,
    };
    for (const child of this._children) {
      const nestedSummary = child.summary;
      summary.alarm += nestedSummary.alarm;
      summary.ok += nestedSummary.ok;
      summary.info += nestedSummary.info;
      summary.skip += nestedSummary.skip;
      summary.error += nestedSummary.error;
    }
    return summary;
  }

  get status(): CheckNodeStatus {
    for (const child of this._children) {
      if (child.status === "error") {
        return "error";
      }
      if (child.status === "unknown") {
        return "unknown";
      }
      if (child.status === "ready") {
        return "ready";
      }
      if (child.status === "started") {
        return "started";
      }
    }
    return "complete";
  }

  get children(): CheckNode[] {
    return this._children;
  }
}

export default RootNode;
