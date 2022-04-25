import {
  CheckNodeStatus,
  CheckNodeType,
  CheckSummary,
  CheckNode,
  CheckResult,
} from "./index";

class ControlResultNode implements CheckNode {
  private readonly _depth: number;
  private readonly _result: CheckResult;

  constructor(depth: number, result: CheckResult) {
    this._depth = depth;
    this._result = result;
  }

  get depth(): number {
    return this._depth;
  }

  get name(): string {
    return this._result.resource;
  }

  get title(): string {
    return this._result.reason;
  }

  get result(): CheckResult {
    return this._result;
  }

  get type(): CheckNodeType {
    return "control_result";
  }

  get summary(): CheckSummary {
    const summary = {
      alarm: 0,
      ok: 0,
      info: 0,
      skip: 0,
      error: 0,
    };
    if (this._result.status === "alarm") {
      summary.alarm += 1;
    }
    if (this._result.status === "error") {
      summary.error += 1;
    }
    if (this._result.status === "ok") {
      summary.ok += 1;
    }
    if (this._result.status === "info") {
      summary.info += 1;
    }
    if (this._result.status === "skip") {
      summary.skip += 1;
    }
    return summary;
  }

  get status(): CheckNodeStatus {
    // for (const child of this._children) {
    //   if (child.status === "error") {
    //     return "error";
    //   }
    //   if (child.status === "unknown") {
    //     return "unknown";
    //   }
    //   if (child.status === "ready") {
    //     return "ready";
    //   }
    //   if (child.status === "started") {
    //     return "started";
    //   }
    // }
    return "complete";
  }
}

export default ControlResultNode;
