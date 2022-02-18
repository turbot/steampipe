import { BasePrimitiveProps } from "../../common";

export interface CheckLeafNodeDataGroupSummaryStatus {
  alarm: number;
  ok: number;
  info: number;
  skip: number;
  error: number;
}

export interface CheckLeafNodeDataGroupSummary {
  status: CheckLeafNodeDataGroupSummaryStatus;
}

interface CheckLeafNodeDataTags {
  [key: string]: string;
}

interface CheckLeafNodeDataControlResultDimensions {
  [key: string]: string;
}

interface CheckLeafNodeDataControlResult {
  reason: string;
  resource: string;
  status: "alarm" | "ok" | "info" | "skip" | "error";
  dimensions: CheckLeafNodeDataControlResultDimensions;
}

export interface CheckLeafNodeDataControl {
  control_id: string;
  title?: string;
  description?: string;
  severity?: string;
  tags?: CheckLeafNodeDataTags;
  results: CheckLeafNodeDataControlResult[];
}

interface CheckLeafNodeDataGroup {
  group_ip: string;
  title?: string;
  description?: string;
  tags?: CheckLeafNodeDataTags;
  summary: CheckLeafNodeDataGroupSummary;
  groups?: CheckLeafNodeDataGroup[];
  controls?: CheckLeafNodeDataControl[];
}

export interface CheckLeafNodeExecutionTree {
  start_time: string;
  end_time: string;
  control_runs: CheckLeafNodeDataControl[];
  root: CheckLeafNodeDataGroup;
}

export interface CheckProps extends BasePrimitiveProps {
  execution_tree?: CheckLeafNodeExecutionTree;
  error?: Error;
}
