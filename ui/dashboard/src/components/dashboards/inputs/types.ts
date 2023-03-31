import { BasePrimitiveProps, ExecutablePrimitiveProps } from "../common";
import { ComponentType } from "react";
import { PanelDefinition } from "../../../types";

export type BaseInputProps = PanelDefinition &
  BasePrimitiveProps &
  ExecutablePrimitiveProps;

export type SelectOption = {
  label: string;
  value: string | null;
  tags?: object;
};

export type SelectInputOption = {
  name: string;
  label?: string;
};

export type InputProperties = {
  label?: string;
  options?: SelectInputOption[];
  placeholder?: string;
  unqualified_name: string;
};

export type InputProps = BaseInputProps & {
  display_type?: InputType;
  properties: InputProperties;
};

export type InputType =
  | "combo"
  | "hidden"
  | "multicombo"
  | "multiselect"
  | "select"
  | "table"
  | "text";

export type IInput = {
  type: InputType;
  component: ComponentType<any>;
};
