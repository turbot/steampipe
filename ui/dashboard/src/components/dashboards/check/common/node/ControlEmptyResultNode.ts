import {
  CheckNodeStatus,
  CheckNodeType,
  CheckSummary,
  CheckNode,
  CheckResult,
  CheckSeveritySummary,
} from "../index";

class ControlEmptyResultNode implements CheckNode {
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
    return "No results";
  }

  get result(): CheckResult {
    return this._result;
  }

  get type(): CheckNodeType {
    return "empty_result";
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
    // If a control has no results, this node is complete
    return "complete";
  }
}

export default ControlEmptyResultNode;
