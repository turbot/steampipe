import { BasePrimitiveProps, ExecutablePrimitiveProps } from "../common";
import React from "react";

export type BaseInputProps = BasePrimitiveProps & ExecutablePrimitiveProps;

interface SelectInputOption {
  name: string;
  label?: string;
}

export type InputProperties = {
  label?: string;
  options?: SelectInputOption[];
  placeholder?: string;
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

export interface IInput {
  type: InputType;
  component: React.ComponentType<any>;
}
