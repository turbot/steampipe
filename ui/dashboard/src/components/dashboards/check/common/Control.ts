import {
  CheckDimensionKeysMap,
  CheckResult,
  CheckRunState,
  CheckSummary,
} from "./index";
import { LeafNodeDataRow } from "../../common";

class Control {
  private readonly _group_id: string;
  private readonly _group_title: string | undefined;
  private readonly _group_description: string | undefined;
  private readonly _name: string;
  private readonly _title: string | undefined;
  private readonly _description: string | undefined;
  private readonly _results: CheckResult[];
  private readonly _summary: CheckSummary;
  private readonly _run_state: CheckRunState;
  private readonly _run_error: string | undefined;

  constructor(
    group_id: string,
    group_title: string | undefined,
    group_description: string | undefined,
    name: string,
    title: string | undefined,
    description: string | undefined,
    results: CheckResult[] | undefined,
    summary: CheckSummary | undefined,
    run_state: number,
    run_error: string | undefined
  ) {
    this._group_id = group_id;
    this._group_title = group_title;
    this._group_description = group_description;
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

  get_dimension_keys(): CheckDimensionKeysMap {
    const dimensionKeysMap = {};
    if (this._results.length === 0) {
      return dimensionKeysMap;
    }
    // for (const result of this._results) {
    //   for (const dimension of result.dimensions) {
    //     dimensionKeysMap[dimension.key] = true;
    //   }
    // }
    for (const dimension of this._results[0].dimensions) {
      dimensionKeysMap[dimension.key] = true;
    }
    return dimensionKeysMap;
  }

  get_data_rows(dimensions: string[]): LeafNodeDataRow[] {
    let rows: LeafNodeDataRow[] = [];
    this._results.forEach((result) => {
      const row: LeafNodeDataRow = [
        this._group_id,
        this._group_title ? this._group_title : null,
        this._group_description ? this._group_description : null,
        this._name,
        this._title ? this._title : null,
        this._description ? this._description : null,
        result.reason,
        result.resource,
        result.status,
      ];

      dimensions.forEach((dimension) => {
        const val = result.dimensions.find((d) => d.key === dimension);
        row.push(val === undefined ? null : val.value);
      });
      rows.push(row);
    });
    return rows;
  }
}

export default Control;
