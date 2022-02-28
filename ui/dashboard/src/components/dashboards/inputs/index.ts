import React from "react";
import MultiSelectInput from "./MultiSelectInput";
import SingleSelectInput from "./SingleSelectInput";
import Table from "../Table";
import { BasePrimitiveProps, ExecutablePrimitiveProps } from "../common";

export type BaseInputProps = BasePrimitiveProps & ExecutablePrimitiveProps;

interface SelectInputOption {
  name: string;
  label?: string;
}

export type InputProperties = {
  type: InputType;
  label?: string;
  options?: SelectInputOption[];
  placeholder?: string;
};

export type InputProps = BaseInputProps & {
  properties: InputProperties;
};

export type InputType = "hidden" | "multiselect" | "select" | "table";

export interface IInput {
  type: InputType;
  component: React.ComponentType<any>;
}

const TableWrapper: IInput = {
  type: "table",
  component: Table,
};

const inputs = {
  [SingleSelectInput.type]: SingleSelectInput,
  [MultiSelectInput.type]: MultiSelectInput,
  [TableWrapper.type]: TableWrapper,
};

export default inputs;
