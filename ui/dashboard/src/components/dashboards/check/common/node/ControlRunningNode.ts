import {
  CheckNodeStatus,
  CheckNodeType,
  CheckSummary,
  CheckNode,
  CheckResult,
  CheckSeveritySummary,
} from "../index";

class ControlRunningNode implements CheckNode {
  private readonly _result: CheckResult;

  constructor(result: CheckResult) {
    this._result = result;
  }

  get sort(): string {
    return this.title;
  }

  get name(): string {
    return this._result.control.name;
  }

  get title(): string {
    return this._result.control.title || this.name;
  }

  get type(): CheckNodeType {
    return "running";
  }

  get summary(): CheckSummary {
    return {
      alarm: 0,
      ok: 0,
      info: 0,
      skip: 0,
      error: 0,
    };
  }

  get severity_summary(): CheckSeveritySummary {
    // Bubble up the node's severity - always zero though as we have no results
    const summary = {};
    if (this._result.control.severity) {
      summary[this._result.control.severity] = 0;
    }
    return summary;
  }

  get status(): CheckNodeStatus {
    // This will bubble up through the hierarchy and put all ancestral nodes in a running state
    return "running";
  }
}

export default ControlRunningNode;
