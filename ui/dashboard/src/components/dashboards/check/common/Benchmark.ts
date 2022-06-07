import Control from "./Control";
import merge from "lodash/merge";
import padStart from "lodash/padStart";
import {
  AddControlResultsAction,
  CheckBenchmarkRun,
  CheckControlRun,
  CheckDynamicColsMap,
  CheckNode,
  CheckNodeStatus,
  CheckNodeType,
  CheckResult,
  CheckSeveritySummary,
  CheckSummary,
} from "./index";
import {
  LeafNodeData,
  LeafNodeDataColumn,
  LeafNodeDataRow,
} from "../../common";

class Benchmark implements CheckNode {
  private readonly _sortIndex: string;
  private readonly _name: string;
  private readonly _title: string;
  private readonly _description?: string;
  private readonly _benchmarks: Benchmark[];
  private readonly _controls: Control[];
  private readonly _add_control_results: AddControlResultsAction;
  private readonly _all_control_results: CheckResult[];

  constructor(
    sortIndex: string,
    name: string,
    title: string | undefined,
    description: string | undefined,
    benchmarks: CheckBenchmarkRun[] | undefined,
    controls: CheckControlRun[] | undefined,
    trunk: Benchmark[],
    add_control_results?: AddControlResultsAction
  ) {
    this._sortIndex = sortIndex;
    this._all_control_results = [];
    this._name = name;
    this._title = title || name;
    this._description = description;

    if (!add_control_results) {
      this._add_control_results = this.add_control_results;
    } else {
      this._add_control_results = add_control_results;
    }

    const thisTrunk = [...trunk, this];
    const nestedBenchmarks: Benchmark[] = [];
    const benchmarksToAdd = benchmarks || [];
    const lengthMaxBenchmarkIndex = (benchmarksToAdd.length - 1).toString()
      .length;
    benchmarksToAdd.forEach((nestedBenchmark, benchmarkIndex) => {
      // @ts-ignore
      const benchmarks = nestedBenchmark.children?.filter(
        (child) => child.node_type === "benchmark_run"
      );
      // @ts-ignore
      const controls = nestedBenchmark.children?.filter(
        (child) => child.node_type === "control_run"
      );
      nestedBenchmarks.push(
        new Benchmark(
          padStart(benchmarkIndex.toString(), lengthMaxBenchmarkIndex),
          nestedBenchmark.name,
          nestedBenchmark.title,
          nestedBenchmark.description,
          benchmarks,
          controls,
          thisTrunk,
          this._add_control_results
        )
      );
    });
    const nestedControls: Control[] = [];
    const controlsToAdd = controls || [];
    const lengthMaxControlIndex = (controlsToAdd.length - 1).toString().length;
    controlsToAdd.forEach((nestedControl, controlIndex) => {
      nestedControls.push(
        new Control(
          padStart(controlIndex.toString(), lengthMaxControlIndex),
          this._name,
          this._title,
          this._description,
          nestedControl.name,
          nestedControl.title,
          nestedControl.description,
          nestedControl.severity,
          nestedControl.results,
          nestedControl.summary,
          nestedControl.tags,
          nestedControl.status,
          nestedControl.error,
          thisTrunk,
          this._add_control_results
        )
      );
    });
    this._benchmarks = nestedBenchmarks;
    this._controls = nestedControls;
  }

  private add_control_results = (results: CheckResult[]) => {
    this._all_control_results.push(...results);
  };

  get all_control_results(): CheckResult[] {
    return this._all_control_results;
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

  get severity_summary(): CheckSeveritySummary {
    return {};
  }

  get status(): CheckNodeStatus {
    for (const benchmark of this._benchmarks) {
      if (benchmark.status === "running") {
        return "running";
      }
    }
    for (const control of this._controls) {
      if (control.status === "running") {
        return "running";
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
        name: "severity",
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
