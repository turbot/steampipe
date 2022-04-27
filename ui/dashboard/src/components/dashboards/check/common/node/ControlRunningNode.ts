import Control from "../Control";
import {
  CheckNodeStatus,
  CheckNodeType,
  CheckSummary,
  CheckNode,
} from "../index";

class ControlRunningNode implements CheckNode {
  private readonly _name: string;
  private readonly _title: string | undefined;

  constructor(control: Control) {
    this._name = `${control.name}-loading`;
    this._title = `${control.name}-loading`;
  }

  get sort(): string {
    return this.title;
  }

  get name(): string {
    return this._name;
  }

  get title(): string {
    return this._title || this._name;
  }

  get type(): CheckNodeType {
    return "running";
  }

  get summary(): CheckSummary {
    return {
      alarm: 0,
      ok: 0,
      info: 0,
      skip: 0,
      error: 0,
    };
  }

  get status(): CheckNodeStatus {
    // This will bubble up through the hierarchy and put all ancestral nodes in a running state
    return "running";
  }
}

export default ControlRunningNode;
