import Control from "./Control";
import { CheckControl, CheckGroup, CheckRunState, CheckSummary } from "./index";

class Benchmark {
  private readonly _name: string;
  private readonly _title?: string;
  private readonly _description?: string;
  private readonly _benchmarks: Benchmark[];
  private readonly _controls: Control[];

  constructor(
    name: string,
    title: string | undefined,
    description: string | undefined,
    benchmarks: CheckGroup[] | undefined,
    controls: CheckControl[] | undefined
  ) {
    this._name = name;
    this._title = title;
    this._description = description;
    const nestedBenchmarks: Benchmark[] = [];
    for (const nestedBenchmark of benchmarks || []) {
      nestedBenchmarks.push(
        new Benchmark(
          nestedBenchmark.group_id,
          nestedBenchmark.title,
          nestedBenchmark.description,
          nestedBenchmark.groups,
          nestedBenchmark.controls
        )
      );
    }
    const nestedControls: Control[] = [];
    for (const nestedControl of controls || []) {
      nestedControls.push(
        new Control(
          nestedControl.control_id,
          nestedControl.title,
          nestedControl.description,
          nestedControl.results,
          nestedControl.summary,
          nestedControl.run_status,
          nestedControl.run_error
        )
      );
    }
    this._benchmarks = nestedBenchmarks;
    this._controls = nestedControls;
    // this.execution_tree = execution_tree;
  }

  get name(): string {
    return this._name;
  }

  get title(): string | undefined {
    return this._title;
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

  get run_state(): CheckRunState {
    for (const benchmark of this._benchmarks) {
      if (benchmark.run_state === "error") {
        return "error";
      }
      if (benchmark.run_state === "unknown") {
        return "unknown";
      }
      if (benchmark.run_state === "ready") {
        return "ready";
      }
      if (benchmark.run_state === "started") {
        return "started";
      }
    }
    for (const control of this._controls) {
      if (control.run_state === "error") {
        return "error";
      }
      if (control.run_state === "unknown") {
        return "unknown";
      }
      if (control.run_state === "ready") {
        return "ready";
      }
      if (control.run_state === "started") {
        return "started";
      }
    }
    return "complete";
  }
}

export default Benchmark;
