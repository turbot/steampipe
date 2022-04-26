import {
  CheckNodeStatus,
  CheckNodeType,
  CheckSummary,
  CheckNode,
} from "../index";

class KeyValuePairNode implements CheckNode {
  private readonly _key: string;
  private readonly _value: string;
  private readonly _children: CheckNode[];

  constructor(key: string, value: string, children?: CheckNode[]) {
    this._key = key;
    this._value = value;
    this._children = children || [];
  }

  get sort(): string {
    return this.title;
  }

  get name(): string {
    return `${this._key}=${this._value}`;
  }

  get title(): string {
    return this._value;
  }

  get type(): CheckNodeType {
    return "dimension";
  }

  get summary(): CheckSummary {
    const summary = {
      alarm: 0,
      ok: 0,
      info: 0,
      skip: 0,
      error: 0,
    };
    for (const child of this._children) {
      const nestedSummary = child.summary;
      summary.alarm += nestedSummary.alarm;
      summary.ok += nestedSummary.ok;
      summary.info += nestedSummary.info;
      summary.skip += nestedSummary.skip;
      summary.error += nestedSummary.error;
    }
    return summary;
  }

  get status(): CheckNodeStatus {
    let hasError = false;
    for (const child of this._children) {
      if (child.status === "ready") {
        return "ready";
      }
      if (child.status === "started") {
        return "started";
      }
      if (child.status === "error") {
        hasError = true;
      }
    }
    return hasError ? "error" : "complete";
  }

  get children(): CheckNode[] {
    return this._children;
  }
}

export default KeyValuePairNode;
