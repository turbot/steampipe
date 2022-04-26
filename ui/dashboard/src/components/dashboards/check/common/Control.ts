import {
  AddControlErrorAction,
  AddControlResultsAction,
  CheckDynamicColsMap,
  CheckNode,
  CheckNodeStatus,
  CheckNodeType,
  CheckResult,
  CheckSummary,
  CheckTags,
} from "./index";
import { LeafNodeDataRow } from "../../common";
import Benchmark from "./Benchmark";

class Control implements CheckNode {
  private readonly _group_id: string;
  private readonly _group_title: string | undefined;
  private readonly _group_description: string | undefined;
  private readonly _name: string;
  private readonly _title: string | undefined;
  private readonly _description: string | undefined;
  private readonly _results: CheckResult[];
  private readonly _summary: CheckSummary;
  private readonly _tags: CheckTags;
  private readonly _run_state: CheckNodeStatus;
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
    tags: CheckTags | undefined,
    status: number,
    run_error: string | undefined,
    benchmark_trunk: Benchmark[],
    add_control_error: AddControlErrorAction,
    add_control_results: AddControlResultsAction
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
    this._tags = tags || {};
    this._run_state = Control._getRunState(status);
    this._run_error = run_error;

    if (this._run_error) {
      add_control_error(this._run_error, this, benchmark_trunk);
    }

    if (this._results) {
      add_control_results(this._results, this, benchmark_trunk);
    }
  }

  private static _getRunState(status: number): CheckNodeStatus {
    if (status === 1) {
      return "ready";
    }
    if (status === 2) {
      return "started";
    }
    if (status === 4) {
      return "complete";
    }
    if (status === 8) {
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

  get type(): CheckNodeType {
    return "control";
  }

  get summary(): CheckSummary {
    return this._summary;
  }

  get error(): string | undefined {
    return this._run_error;
  }

  get status(): CheckNodeStatus {
    return this._run_state;
  }

  get results(): CheckResult[] {
    return this._results;
  }

  get tags(): CheckTags {
    return this._tags;
  }

  get_dynamic_cols(): CheckDynamicColsMap {
    const dimensionKeysMap = {
      dimensions: {},
      tags: {},
    };

    Object.keys(this._tags).forEach((t) => (dimensionKeysMap.tags[t] = true));

    if (this._results.length === 0) {
      return dimensionKeysMap;
    }
    // for (const result of this._results) {
    //   for (const dimension of result.dimensions) {
    //     dimensionKeysMap[dimension.key] = true;
    //   }
    // }
    for (const dimension of this._results[0].dimensions) {
      dimensionKeysMap.dimensions[dimension.key] = true;
    }
    return dimensionKeysMap;
  }

  get_data_rows(tags: string[], dimensions: string[]): LeafNodeDataRow[] {
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

      tags.forEach((tag) => {
        const val = this._tags[tag];
        row.push(val === undefined ? null : val);
      });

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
