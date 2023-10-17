import Benchmark from "./Benchmark";
import {
  BasePrimitiveProps,
  ExecutablePrimitiveProps,
  LeafNodeData,
} from "../../common";
import { DashboardRunState } from "../../../../types";

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

export type CheckNode = {
  sort: string;
  name: string;
  title: string;
  type: CheckNodeType;
  status: CheckNodeStatus;
  severity?: CheckSeverity;
  severity_summary: CheckSeveritySummary;
  summary: CheckSummary;
  children?: CheckNode[];
  data?: LeafNodeData;
  error?: string;
  merge?: (other: CheckNode) => void;
};

export type CheckNodeStatus = "running" | "complete";

export type CheckSeverity = "none" | "low" | "medium" | "high" | "critical";

export type CheckSeveritySummary =
  | {}
  | {
      [key in CheckSeverity]: number;
    };

export type CheckSummary = {
  alarm: number;
  ok: number;
  info: number;
  skip: number;
  error: number;
};

export type CheckDynamicValueMap = {
  [dimension: string]: boolean;
};

export type CheckDynamicColsMap = {
  dimensions: CheckDynamicValueMap;
  tags: CheckDynamicValueMap;
};

export type CheckTags = {
  [key: string]: string;
};

export type CheckResultDimension = {
  key: string;
  value: string;
};

export enum CheckResultStatus {
  alarm = 0,
  ok = 1,
  info = 2,
  skip = 3,
  error = 4,
  empty = 5,
}

export type CheckResultType = "loading" | "error" | "empty" | "result";

export type CheckResult = {
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
};

type CheckControlRunProperties = {
  severity?: CheckSeverity | undefined;
};

export type CheckControlRun = {
  name: string;
  title?: string;
  description?: string;
  panel_type: "control";
  properties?: CheckControlRunProperties;
  severity?: CheckSeverity | undefined;
  tags?: CheckTags;
  data: LeafNodeData;
  summary: CheckSummary;
  status: DashboardRunState;
  error?: string;
};

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

export type CheckDisplayGroup = {
  id: string;
  type: CheckDisplayGroupType;
  value?: string | undefined;
};

export interface Filter {
  type: CheckFilterType;
  key?: string;
  value?: string;
}

export interface OrFilter {
  or?: Filter[];
}

export interface AndFilter {
  and?: (Filter | OrFilter)[];
}

export type CheckFilter = AndFilter & OrFilter;

export type CheckFilterType = CheckDisplayGroupType;

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
  key?: string,
) => {
  if (!dimensions || !key) {
    return undefined;
  }
  return dimensions.find((d) => d.key === key);
};
