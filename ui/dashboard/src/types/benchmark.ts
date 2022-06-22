import { PanelDefinition } from "./panel";

export interface BenchmarkDefinition extends PanelDefinition {
  children?: BenchmarkDefinition | ControlDefinition[];
}

export interface ControlDefinition extends PanelDefinition {}
