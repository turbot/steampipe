import {
  CheckNodeStatus,
  CheckNodeType,
  CheckSummary,
  CheckNode,
  CheckResult,
} from "./index";

class ControlNode implements CheckNode {
  private readonly _depth: number;
  private readonly _name: string;
  private readonly _title: string | undefined;
  private readonly _children: CheckNode[];
  private readonly _results: CheckResult[];

  constructor(
    depth: number,
    name: string,
    title: string | undefined,
    children?: CheckNode[]
  ) {
    this._depth = depth;
    this._name = name;
    this._title = title;
    this._children = children || [];
    this._results = [];
  }

  get depth(): number {
    return this._depth;
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
    for (const result of this._results) {
      if (result.status === "alarm") {
        summary.alarm += 1;
      }
      if (result.status === "error") {
        summary.error += 1;
      }
      if (result.status === "ok") {
        summary.ok += 1;
      }
      if (result.status === "info") {
        summary.info += 1;
      }
      if (result.status === "skip") {
        summary.skip += 1;
      }
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

  get results(): CheckResult[] {
    return this._results;
  }
}

export default ControlNode;
