import Benchmark from "./Benchmark";
import { BasePrimitiveProps, ExecutablePrimitiveProps } from "../../common";

export type CheckNodeType = "benchmark" | "control";

export interface CheckNode {
  depth: number;
  name: string;
  title?: string;
  type: CheckNodeType;
  status: CheckNodeStatus;
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

export type CheckNodeStatus =
  | "ready"
  | "started"
  | "complete"
  | "error"
  | "unknown";

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
  reason: string;
  resource: string;
  status: CheckResultStatus;
  dimensions: CheckResultDimension[];
}

export interface CheckControl {
  control_id: string;
  title?: string;
  description?: string;
  severity?: string;
  tags?: CheckTags;
  results: CheckResult[];
  summary: CheckSummary;
  run_status: number;
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

export type CheckProps = BasePrimitiveProps &
  ExecutablePrimitiveProps & {
    type?: CheckType;
    properties: {
      display: "all" | "none";
      type?: CheckType;
    };
  };

export type BenchmarkTreeProps = BasePrimitiveProps &
  ExecutablePrimitiveProps & {
    properties: {
      benchmark: Benchmark;
    };
  };
