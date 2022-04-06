import { CheckResult, CheckRunState, CheckSummary } from "./index";

class Control {
  private readonly _name: string;
  private readonly _title: string | undefined;
  private readonly _description: string | undefined;
  private readonly _results: CheckResult[] | undefined;
  private readonly _summary: CheckSummary;
  private readonly _run_state: CheckRunState;

  constructor(
    name: string,
    title: string | undefined,
    description: string | undefined,
    results: CheckResult[] | undefined,
    summary: CheckSummary | undefined,
    run_state: number
  ) {
    this._name = name;
    this._title = title;
    this._description = description;
    this._results = results;
    this._summary = summary || {
      alarm: 0,
      ok: 0,
      info: 0,
      skip: 0,
      error: 0,
    };
    this._run_state = Control._getRunState(run_state);
  }

  private static _getRunState(run_state: number): CheckRunState {
    if (run_state === 1) {
      return "ready";
    }
    if (run_state === 2) {
      return "started";
    }
    if (run_state === 4) {
      return "complete";
    }
    if (run_state === 8) {
      return "error";
    }
    return "unknown";
  }

  get name(): string {
    return this._name;
  }

  get title(): string | undefined {
    return this._title;
  }

  get summary(): CheckSummary {
    return this._summary;
  }

  get run_state(): CheckRunState {
    return this._run_state;
  }
}

export default Control;
