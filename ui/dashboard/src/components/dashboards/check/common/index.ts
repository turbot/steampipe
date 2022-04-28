import Control from "./Control";
import { BasePrimitiveProps, ExecutablePrimitiveProps } from "../../common";
import Benchmark from "./Benchmark";

export type CheckNodeType =
  | "benchmark"
  | "control"
  | "dimension"
  | "error"
  | "reason"
  | "resource"
  | "result"
  | "running"
  | "root"
  | "severity"
  | "status"
  | "tag";

export interface CheckNode {
  sort: string;
  name: string;
  title: string;
  type: CheckNodeType;
  status: CheckNodeStatus;
  severity?: CheckSeverity;
  severity_summary: CheckSeveritySummary;
  summary: CheckSummary;
  children?: CheckNode[];
  results?: CheckResult[];
  error?: string;
}

// export interface IControl {
//   name: string;
//   title: string;
//   description?: string;
//   summary: CheckSummary;
//   results?: IControlResult[];
// }

// export interface IBenchmark {
//   name: string;
//   title: string;
//   description?: string;
//   summary: CheckSummary;
//   benchmarks: IBenchmark[];
//   controls: IControl[];
// }

export type CheckNodeStatusRaw = 1 | 2 | 4 | 8;

export type CheckNodeStatus = "running" | "complete";

export type CheckSeverity = "none" | "low" | "medium" | "high" | "critical";

export type CheckSeveritySummary =
  | {}
  | {
      [key in CheckSeverity]: number;
    };

export interface CheckSummary {
  alarm: number;
  ok: number;
  info: number;
  skip: number;
  error: number;
}
export interface CheckLeafNodeDataGroupSummary {
  status: CheckSummary;
}

export interface CheckDynamicValueMap {
  [dimension: string]: boolean;
}

export interface CheckDynamicColsMap {
  dimensions: CheckDynamicValueMap;
  tags: CheckDynamicValueMap;
}

export interface CheckTags {
  [key: string]: string;
}

interface CheckResultDimension {
  key: string;
  value: string;
}

export type CheckResultStatus = "alarm" | "ok" | "info" | "skip" | "error";

export interface CheckResult {
  dimensions: CheckResultDimension[];
  tags: CheckTags;
  control: CheckNode;
  benchmark_trunk: Benchmark[];
  status: CheckResultStatus;
  reason: string;
  resource: string;
  severity?: CheckSeverity;
  error?: string;
}

export interface CheckControl {
  control_id: string;
  title?: string;
  description?: string;
  severity?: CheckSeverity | undefined;
  tags?: CheckTags;
  results: CheckResult[];
  summary: CheckSummary;
  run_status: CheckNodeStatusRaw;
  run_error?: string;
}

export interface CheckGroup {
  group_id: string;
  title?: string;
  description?: string;
  tags?: CheckTags;
  summary: CheckLeafNodeDataGroupSummary;
  groups?: CheckGroup[];
  controls?: CheckControl[];
}

interface CheckLeafNodeProgress {
  summary: CheckSummary;
}

export interface CheckExecutionTree {
  start_time: string;
  end_time: string;
  control_runs: CheckControl[];
  progress: CheckLeafNodeProgress;
  root: CheckGroup;
}

type CheckType = "summary" | "table" | null;

export interface CheckDisplayGroup {
  type:
    | "benchmark"
    | "control"
    | "result"
    | "tag"
    | "dimension"
    | "reason"
    | "resource"
    | "severity"
    | "status";
  value?: string;
}

export type CheckProps = BasePrimitiveProps & ExecutablePrimitiveProps;

export type BenchmarkTreeProps = BasePrimitiveProps &
  ExecutablePrimitiveProps & {
    properties: {
      grouping: CheckNode;
      grouping_config: CheckDisplayGroup[];
      first_child_summaries: CheckSummary[];
    };
  };

export type AddControlLoadingAction = (
  benchmark_trunk: Benchmark[],
  control: Control
) => void;

export type AddControlErrorAction = (
  error: string,
  benchmark_trunk: Benchmark[],
  control: Control
) => void;

export type AddControlResultsAction = (
  results: CheckResult[],
  benchmark_trunk: Benchmark[],
  control: Control
) => void;
