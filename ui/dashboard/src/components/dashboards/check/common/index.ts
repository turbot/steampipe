import { BasePrimitiveProps } from "../../common";

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

export type CheckRunState =
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

export interface CheckDimensionKeysMap {
  [dimension: string]: boolean;
}

interface CheckTags {
  [key: string]: string;
}

interface CheckResultDimension {
  key: string;
  value: string;
}

export interface CheckResult {
  reason: string;
  resource: string;
  status: "alarm" | "ok" | "info" | "skip" | "error";
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

export interface CheckProps extends BasePrimitiveProps {
  root: CheckGroup;
  error?: Error;
  properties: {
    type?: CheckType;
  };
}
