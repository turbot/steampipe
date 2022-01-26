import React from "react";
import SelectInput from "./SelectInput";
import Table from "../Table";
import { BasePrimitiveProps, ExecutablePrimitiveProps } from "../common";

export type BaseInputProps = BasePrimitiveProps & ExecutablePrimitiveProps;

export type InputProperties = {
  type: InputType;
};

export type InputProps = BaseInputProps & {
  properties: InputProperties;
};

export type InputType = "select" | "table";

export interface IInput {
  type: InputType;
  component: React.ComponentType<any>;
}

const TableWrapper: IInput = {
  type: "table",
  component: Table,
};

const inputs = {
  [SelectInput.type]: SelectInput,
  [TableWrapper.type]: TableWrapper,
};

export default inputs;
