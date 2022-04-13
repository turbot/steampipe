import { CheckResult, CheckRunState, CheckSummary } from "./index";

class Control {
  private readonly _name: string;
  private readonly _title: string | undefined;
  private readonly _description: string | undefined;
  private readonly _results: CheckResult[];
  private readonly _summary: CheckSummary;
  private readonly _run_state: CheckRunState;
  private readonly _run_error: string | undefined;

  constructor(
    name: string,
    title: string | undefined,
    description: string | undefined,
    results: CheckResult[] | undefined,
    summary: CheckSummary | undefined,
    run_state: number,
    run_error: string | undefined
  ) {
    this._name = name;
    this._title = title;
    this._description = description;
    this._results = results || [];
    this._summary = summary || {
      alarm: 0,
      ok: 0,
      info: 0,
      skip: 0,
      error: 0,
    };
    this._run_state = Control._getRunState(run_state);
    this._run_error = run_error;
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

  get run_error(): string | undefined {
    return this._run_error;
  }

  get run_state(): CheckRunState {
    return this._run_state;
  }

  get results(): CheckResult[] {
    return this._results;
  }
}

export default Control;
