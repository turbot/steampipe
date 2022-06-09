import Benchmark from "./Benchmark";
import {
  AddControlResultsAction,
  CheckDynamicColsMap,
  CheckNode,
  CheckNodeStatus,
  CheckNodeStatusRaw,
  CheckNodeType,
  CheckResult,
  CheckSeverity,
  CheckSeveritySummary,
  CheckSummary,
  CheckTags,
  findDimension,
} from "./index";
import { LeafNodeDataRow } from "../../common";
import { PanelsMap } from "../../../../hooks/useDashboard";

class Control implements CheckNode {
  private readonly _sortIndex: string;
  private readonly _group_id: string;
  private readonly _group_title: string | undefined;
  private readonly _group_description: string | undefined;
  private readonly _name: string;
  private readonly _title: string | undefined;
  private readonly _description: string | undefined;
  private readonly _severity: CheckSeverity | undefined;
  private readonly _results: CheckResult[];
  private readonly _summary: CheckSummary;
  private readonly _tags: CheckTags;
  private readonly _status: CheckNodeStatusRaw;
  private readonly _error: string | undefined;

  constructor(
    sortIndex: string,
    group_id: string,
    group_title: string | undefined,
    group_description: string | undefined,
    name: string,
    title: string | undefined,
    description: string | undefined,
    severity: CheckSeverity | undefined,
    results: CheckResult[] | undefined,
    summary: CheckSummary | undefined,
    tags: CheckTags | undefined,
    status: CheckNodeStatusRaw,
    error: string | undefined,
    panelsMap: PanelsMap,
    benchmark_trunk: Benchmark[],
    add_control_results: AddControlResultsAction
  ) {
    this._sortIndex = sortIndex;
    this._group_id = group_id;
    this._group_title = group_title;
    this._group_description = group_description;
    this._name = name;
    this._title = title;
    this._description = description;
    this._severity = severity;
    this._results = results || [];
    this._summary = summary || {
      alarm: 0,
      ok: 0,
      info: 0,
      skip: 0,
      error: 0,
    };
    this._tags = tags || {};
    this._status = status;
    this._error = error;

    if (this._status === "ready" || this._status === "started") {
      add_control_results([this._build_control_loading_node(benchmark_trunk)]);
    } else if (this._error) {
      add_control_results([
        this._build_control_error_node(benchmark_trunk, this._error),
      ]);
    } else if (!this._results || this._results.length === 0) {
      add_control_results([this._build_control_empty_result(benchmark_trunk)]);
    } else {
      add_control_results(
        this._build_control_results(benchmark_trunk, this._results)
      );
    }
  }

  get sort(): string {
    return `${this._sortIndex}-${this.title}`;
  }

  get name(): string {
    return this._name;
  }

  get title(): string {
    return this._title || this._name;
  }

  get severity(): CheckSeverity | undefined {
    return this._severity;
  }

  get severity_summary(): CheckSeveritySummary {
    return {};
  }

  get type(): CheckNodeType {
    return "control";
  }

  get summary(): CheckSummary {
    return this._summary;
  }

  get error(): string | undefined {
    return this._error;
  }

  get status(): CheckNodeStatus {
    switch (this._status) {
      case "ready":
      case "started":
        return "running";
      default:
        return "complete";
    }
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
    for (const result of this._results) {
      for (const dimension of result.dimensions || []) {
        dimensionKeysMap.dimensions[dimension.key] = true;
      }
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
        this._severity ? this._severity : null,
        result.reason,
        result.resource,
        result.status,
      ];

      tags.forEach((tag) => {
        const val = this._tags[tag];
        row.push(val === undefined ? null : val);
      });

      dimensions.forEach((dimension) => {
        const val = findDimension(result.dimensions, dimension);
        row.push(val === undefined ? null : val.value);
      });

      rows.push(row);
    });
    return rows;
  }

  private _build_control_loading_node = (
    benchmark_trunk: Benchmark[]
  ): CheckResult => {
    return {
      type: "loading",
      dimensions: [],
      tags: this.tags,
      control: this,
      reason: "",
      resource: "",
      status: "ok",
      benchmark_trunk,
    };
  };

  private _build_control_error_node = (
    benchmark_trunk: Benchmark[],
    error: string
  ): CheckResult => {
    return {
      type: "error",
      error,
      dimensions: [],
      tags: this.tags,
      control: this,
      reason: "",
      resource: "",
      status: "error",
      benchmark_trunk,
    };
  };

  private _build_control_empty_result = (
    benchmark_trunk: Benchmark[]
  ): CheckResult => {
    return {
      type: "empty",
      error: undefined,
      dimensions: [],
      tags: this.tags,
      control: this,
      reason: "",
      resource: "",
      status: "empty",
      benchmark_trunk,
    };
  };

  private _build_control_results = (
    benchmark_trunk: Benchmark[],
    results: CheckResult[]
  ): CheckResult[] => {
    return results.map((r) => ({
      ...r,
      type: "result",
      severity: this.severity,
      tags: this.tags,
      benchmark_trunk,
      control: this,
    }));
  };
}

export default Control;
