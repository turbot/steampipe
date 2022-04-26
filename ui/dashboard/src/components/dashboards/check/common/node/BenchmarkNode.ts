import {
  CheckNodeStatus,
  CheckNodeType,
  CheckSummary,
  CheckNode,
} from "../index";

class BenchmarkNode implements CheckNode {
  private readonly _sort: string;
  private readonly _name: string;
  private readonly _title: string | undefined;
  private readonly _children: CheckNode[];

  constructor(
    sort: string,
    name: string,
    title: string | undefined,
    children?: CheckNode[]
  ) {
    this._sort = sort;
    this._name = name;
    this._title = title;
    this._children = children || [];
  }

  get sort(): string {
    return this._sort;
  }

  get name(): string {
    return this._name;
  }

  get title(): string {
    return this._title || this._name;
  }

  get type(): CheckNodeType {
    return "dimension";
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
    let hasError = false;
    for (const child of this._children) {
      if (child.status === "ready") {
        return "ready";
      }
      if (child.status === "started") {
        return "started";
      }
      if (child.status === "error") {
        hasError = true;
      }
    }
    return hasError ? "error" : "complete";
  }

  get children(): CheckNode[] {
    return this._children;
  }
}

export default BenchmarkNode;
