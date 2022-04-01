import { CheckResult, CheckSummary } from "./index";

class Control {
  private readonly _name: string;
  private readonly _title: string | undefined;
  private readonly _description: string | undefined;
  private readonly _results: CheckResult[] | undefined;
  private readonly _summary: CheckSummary;

  constructor(
    name: string,
    title: string | undefined,
    description: string | undefined,
    results: CheckResult[] | undefined,
    summary: CheckSummary | undefined
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
}

export default Control;
