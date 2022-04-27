import {
  CheckNodeStatus,
  CheckNodeType,
  CheckSummary,
  CheckNode,
  CheckResult,
  CheckSeverity,
  CheckSeveritySummary,
} from "../index";

class ControlErrorNode implements CheckNode {
  private readonly _name: string;
  private readonly _title: string | undefined;
  private readonly _error: string | undefined;

  constructor(check_error: CheckResult) {
    this._name = check_error.control.name;
    this._title = check_error.control.title;
    this._error = check_error.error;
  }

  get sort(): string {
    return this.title;
  }

  get name(): string {
    return this._name;
  }

  get title(): string {
    return this._title || this._name;
  }

  get error(): string {
    return this._error || "Unknown error";
  }

  get type(): CheckNodeType {
    return "error";
  }

  get summary(): CheckSummary {
    return {
      alarm: 0,
      ok: 0,
      info: 0,
      skip: 0,
      error: 1,
    };
  }

  get severity_summary(): CheckSeveritySummary {
    return {};
  }

  get status(): CheckNodeStatus {
    // If a control has gone to error, this node is complete
    return "complete";
  }
}

export default ControlErrorNode;
