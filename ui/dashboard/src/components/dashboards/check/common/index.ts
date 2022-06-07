import Benchmark from "./Benchmark";
import { BasePrimitiveProps, ExecutablePrimitiveProps } from "../../common";

export type CheckNodeType =
  | "benchmark"
  | "control"
  | "dimension"
  | "empty_result"
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
  merge?: (other: CheckNode) => void;
}

export type CheckNodeStatusRaw = "ready" | "started" | "complete" | "error";

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

export type CheckResultStatus =
  | "alarm"
  | "ok"
  | "info"
  | "skip"
  | "error"
  | "empty";

export type CheckResultType = "loading" | "error" | "empty" | "result";

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
  type: CheckResultType;
}

export interface CheckControlRun {
  name: string;
  title?: string;
  description?: string;
  node_type: "control_run";
  severity?: CheckSeverity | undefined;
  tags?: CheckTags;
  results: CheckResult[];
  summary: CheckSummary;
  status: CheckNodeStatusRaw;
  error?: string;
}

export type CheckGroupType = "benchmark" | "table";

export interface CheckBenchmarkRun {
  name: string;
  title?: string;
  description?: string;
  node_type: "benchmark";
  tags?: CheckTags;
  type: CheckGroupType;
  summary: CheckLeafNodeDataGroupSummary;
  children?: CheckBenchmarkRun[] | CheckControlRun[];
}

interface CheckLeafNodeProgress {
  summary: CheckSummary;
}

export interface CheckExecutionTree {
  start_time: string;
  end_time: string;
  control_runs: CheckControlRun[];
  progress: CheckLeafNodeProgress;
  root: CheckBenchmarkRun;
}

export type CheckDisplayGroupType =
  | "benchmark"
  | "control"
  | "result"
  | "tag"
  | "dimension"
  | "reason"
  | "resource"
  | "severity"
  | "status";

export interface CheckDisplayGroup {
  type: CheckDisplayGroupType;
  value?: string;
}

export type BenchmarkTreeProps = BasePrimitiveProps &
  ExecutablePrimitiveProps & {
    properties: {
      grouping: CheckNode;
      first_child_summaries: CheckSummary[];
    };
  };

export type AddControlResultsAction = (results: CheckResult[]) => void;

export const findDimension = (
  dimensions?: CheckResultDimension[],
  key?: string
) => {
  if (!dimensions || !key) {
    return undefined;
  }
  return dimensions.find((d) => d.key === key);
};
