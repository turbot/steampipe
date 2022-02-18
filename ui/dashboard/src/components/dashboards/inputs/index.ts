import React from "react";
import MultiSelectInput from "./MultiSelectInput";
import SingleSelectInput from "./SingleSelectInput";
import Table from "../Table";
import { BasePrimitiveProps, ExecutablePrimitiveProps } from "../common";

export type BaseInputProps = BasePrimitiveProps & ExecutablePrimitiveProps;

export type InputProperties = {
  type: InputType;
  placeholder?: string;
};

export type InputProps = BaseInputProps & {
  properties: InputProperties;
};

export type InputType = "multi" | "select" | "table";

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
