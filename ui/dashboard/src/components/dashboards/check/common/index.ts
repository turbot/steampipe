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
  control_row_status_summary: CheckSummary;
}

export interface CheckGroup {
  group_ip: string;
  title?: string;
  description?: string;
  tags?: CheckTags;
  summary: CheckLeafNodeDataGroupSummary;
  child_groups?: CheckGroup[];
  child_controls?: CheckControl[];
}

interface CheckLeafNodeProgress {
  control_row_status_summary: CheckSummary;
}

export interface CheckExecutionTree {
  start_time: string;
  end_time: string;
  control_runs: CheckControl[];
  progress: CheckLeafNodeProgress;
  root: CheckGroup;
}

export interface CheckProps extends BasePrimitiveProps {
  root: CheckGroup;
  error?: Error;
}
