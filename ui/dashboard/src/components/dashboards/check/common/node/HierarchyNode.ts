import {
  CheckNode,
  CheckNodeStatus,
  CheckNodeType,
  CheckSeveritySummary,
  CheckSummary,
} from "../index";

class HierarchyNode implements CheckNode {
  private readonly _type: CheckNodeType;
  private readonly _name: string;
  private readonly _title: string;
  private readonly _sort: string;
  private readonly _children: CheckNode[];

  constructor(
    type: CheckNodeType,
    name: string,
    title: string,
    sort: string,
    children: CheckNode[]
  ) {
    this._type = type;
    this._name = name;
    this._title = title;
    this._sort = sort;
    this._children = children;
  }

  get type(): CheckNodeType {
    return this._type;
  }

  get name(): string {
    return this._name;
  }

  get title(): string {
    return this._title;
  }

  get sort(): string {
    return this._sort;
  }

  get children(): CheckNode[] {
    return this._children;
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

  get severity_summary(): CheckSeveritySummary {
    const summary = {};
    for (const child of this._children) {
      for (const [severity, count] of Object.entries(child.severity_summary)) {
        if (!summary[severity]) {
          summary[severity] = count;
        } else {
          summary[severity] += count;
        }
      }
    }
    return summary;
  }

  get status(): CheckNodeStatus {
    for (const child of this._children) {
      if (child.status === "running") {
        return "running";
      }
    }
    return "complete";
  }

  merge(other: CheckNode) {
    // merge(other) -> iterate children of other -> if child exists on me, call me_child.merge(other_child), else add to end of children
    for (const otherChild of other.children || []) {
      // Check for existing child with this name
      const matchingSelfChild = this.children.find(
        (selfChild) => selfChild.name === otherChild.name
      );

      if (matchingSelfChild) {
        if (!matchingSelfChild.merge) {
          continue;
        }
        // If there's a matching child, merge that child in
        matchingSelfChild.merge(otherChild);
      } else {
        // Else append to my children
        this.children.push(otherChild);
      }
    }
  }
}

export default HierarchyNode;
