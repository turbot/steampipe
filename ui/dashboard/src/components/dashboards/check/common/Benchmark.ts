import Control from "./Control";
import merge from "lodash/merge";
import {
  AddControlErrorAction,
  AddControlResultsAction,
  CheckControl,
  CheckDynamicColsMap,
  CheckError,
  CheckGroup,
  CheckNode,
  CheckNodeStatus,
  CheckNodeType,
  CheckResult,
  CheckSummary,
} from "./index";
import {
  LeafNodeData,
  LeafNodeDataColumn,
  LeafNodeDataRow,
} from "../../common";

class Benchmark implements CheckNode {
  private readonly _name: string;
  private readonly _title: string;
  private readonly _description?: string;
  private readonly _benchmarks: Benchmark[];
  private readonly _controls: Control[];
  private readonly _add_control_error: AddControlErrorAction;
  private readonly _add_control_results: AddControlResultsAction;
  private readonly _all_control_errors: CheckError[];
  private readonly _all_control_results: CheckResult[];

  constructor(
    name: string,
    title: string | undefined,
    description: string | undefined,
    benchmarks: CheckGroup[] | undefined,
    controls: CheckControl[] | undefined,
    trunk: Benchmark[],
    add_control_error?: AddControlErrorAction,
    add_control_results?: AddControlResultsAction
  ) {
    this._all_control_errors = [];
    this._all_control_results = [];
    this._name = name;
    this._title = title || name;
    this._description = description;

    if (!add_control_error) {
      this._add_control_error = this.add_control_error;
    } else {
      this._add_control_error = add_control_error;
    }

    if (!add_control_results) {
      this._add_control_results = this.add_control_results;
    } else {
      this._add_control_results = add_control_results;
    }

    const nestedBenchmarks: Benchmark[] = [];
    for (const nestedBenchmark of benchmarks || []) {
      nestedBenchmarks.push(
        new Benchmark(
          nestedBenchmark.group_id,
          nestedBenchmark.title,
          nestedBenchmark.description,
          nestedBenchmark.groups,
          nestedBenchmark.controls,
          [...trunk, this],
          this._add_control_error,
          this._add_control_results
        )
      );
    }
    const nestedControls: Control[] = [];
    for (const nestedControl of controls || []) {
      nestedControls.push(
        new Control(
          this._name,
          this._title,
          this._description,
          nestedControl.control_id,
          nestedControl.title,
          nestedControl.description,
          nestedControl.results,
          nestedControl.summary,
          nestedControl.tags,
          nestedControl.run_status,
          nestedControl.run_error,
          [...trunk, this],
          this._add_control_error,
          this._add_control_results
        )
      );
    }
    this._benchmarks = nestedBenchmarks;
    this._controls = nestedControls;
  }

  private add_control_error = (
    error: string,
    control: Control,
    benchmark_trunk: Benchmark[]
  ) => {
    this._all_control_errors.push({
      error,
      dimensions: [],
      tags: control.tags,
      control,
      benchmark_trunk,
    });
  };

  private add_control_results = (
    results: CheckResult[],
    control: Control,
    benchmark_trunk: Benchmark[]
  ) => {
    this._all_control_results.push(
      ...results.map((r) => ({
        ...r,
        tags: control.tags,
        control,
        benchmark_trunk,
      }))
    );
  };

  get all_control_errors(): CheckError[] {
    return this._all_control_errors;
  }

  get all_control_results(): CheckResult[] {
    return this._all_control_results;
  }

  get name(): string {
    return this._name;
  }

  get title(): string {
    return this._title || this._name;
  }

  get type(): CheckNodeType {
    return "benchmark";
  }

  get children(): CheckNode[] {
    return [...this._benchmarks, ...this._controls];
  }

  get benchmarks(): Benchmark[] {
    return this._benchmarks;
  }

  get controls(): Control[] {
    return this._controls;
  }

  get summary(): CheckSummary {
    const summary = {
      alarm: 0,
      ok: 0,
      info: 0,
      skip: 0,
      error: 0,
    };
    for (const benchmark of this._benchmarks) {
      const nestedSummary = benchmark.summary;
      summary.alarm += nestedSummary.alarm;
      summary.ok += nestedSummary.ok;
      summary.info += nestedSummary.info;
      summary.skip += nestedSummary.skip;
      summary.error += nestedSummary.error;
    }
    for (const control of this._controls) {
      const nestedSummary = control.summary;
      summary.alarm += nestedSummary.alarm;
      summary.ok += nestedSummary.ok;
      summary.info += nestedSummary.info;
      summary.skip += nestedSummary.skip;
      summary.error += nestedSummary.error;
    }
    return summary;
  }

  get status(): CheckNodeStatus {
    for (const benchmark of this._benchmarks) {
      if (benchmark.status === "error") {
        return "error";
      }
      if (benchmark.status === "unknown") {
        return "unknown";
      }
      if (benchmark.status === "ready") {
        return "ready";
      }
      if (benchmark.status === "started") {
        return "started";
      }
    }
    for (const control of this._controls) {
      if (control.status === "error") {
        return "error";
      }
      if (control.status === "unknown") {
        return "unknown";
      }
      if (control.status === "ready") {
        return "ready";
      }
      if (control.status === "started") {
        return "started";
      }
    }
    return "complete";
  }

  get_data_table(): LeafNodeData {
    const columns: LeafNodeDataColumn[] = [
      {
        name: "group_id",
        data_type_name: "TEXT",
      },
      {
        name: "title",
        data_type_name: "TEXT",
      },
      {
        name: "description",
        data_type_name: "TEXT",
      },
      {
        name: "control_id",
        data_type_name: "TEXT",
      },
      {
        name: "control_title",
        data_type_name: "TEXT",
      },
      {
        name: "control_description",
        data_type_name: "TEXT",
      },
      {
        name: "reason",
        data_type_name: "TEXT",
      },
      {
        name: "resource",
        data_type_name: "TEXT",
      },
      {
        name: "status",
        data_type_name: "TEXT",
      },
    ];
    const { dimensions, tags } = this.get_dynamic_cols();
    Object.keys(tags).forEach((tag) =>
      columns.push({
        name: tag,
        data_type_name: "TEXT",
      })
    );
    Object.keys(dimensions).forEach((dimension) =>
      columns.push({
        name: dimension,
        data_type_name: "TEXT",
      })
    );
    const rows = this.get_data_rows(Object.keys(tags), Object.keys(dimensions));
    // let rows: LeafNodeDataRow[] = [];
    // this._benchmarks.forEach(benchmark => {
    //   rows = [...rows, ...benchmark.get_data_rows()]
    // })
    // this._controls.forEach(control => {
    //   rows = [...rows, ...control.get_data_rows()]
    // })

    return {
      columns,
      rows,
    };
  }

  get_dynamic_cols(): CheckDynamicColsMap {
    let keys = {
      dimensions: {},
      tags: {},
    };
    this._benchmarks.forEach((benchmark) => {
      const subBenchmarkKeys = benchmark.get_dynamic_cols();
      keys = merge(keys, subBenchmarkKeys);
    });
    this._controls.forEach((control) => {
      const controlKeys = control.get_dynamic_cols();
      keys = merge(keys, controlKeys);
    });
    return keys;
  }

  get_data_rows(tags: string[], dimensions: string[]): LeafNodeDataRow[] {
    let rows: LeafNodeDataRow[] = [];
    this._benchmarks.forEach((benchmark) => {
      rows = [...rows, ...benchmark.get_data_rows(tags, dimensions)];
    });
    this._controls.forEach((control) => {
      rows = [...rows, ...control.get_data_rows(tags, dimensions)];
    });
    return rows;
  }
}

export default Benchmark;
