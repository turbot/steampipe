import IntegerDisplay from "../../IntegerDisplay";
import LoadingIndicator from "../LoadingIndicator";
import React, { useEffect, useState } from "react";
import Table from "../Table";
import { AlertIcon, InfoIcon, NilIcon, OKIcon } from "../../../constants/icons";
import {
  BasePrimitiveProps,
  ExecutablePrimitiveProps,
  LeafNodeData,
} from "../common";
import { get } from "lodash";
import { getColumnIndex } from "../../../utils/data";

const getWrapperClasses = (type) => {
  switch (type) {
    case "alert":
      return "bg-alert";
    case "info":
      return "bg-info";
    case "ok":
      return "bg-ok";
    default:
      return "bg-black-scale-2";
  }
};

const getTextClasses = (type) => {
  switch (type) {
    case "alert":
      return "text-alert-inverse";
    case "info":
      return "text-info-inverse";
    case "ok":
      return "text-ok-inverse";
    default:
      return null;
  }
};

export type CardType = "alert" | "info" | "ok" | "table" | null;

export type CardProps = BasePrimitiveProps &
  ExecutablePrimitiveProps & {
    properties: {
      type?: CardType;
      value?: string;
    };
  };

type CardDataFormat = "simple" | "formal";

interface CardState {
  loading: boolean;
  label: string | null;
  value: number | null;
  type: CardType;
}

const getDataFormat = (data: LeafNodeData): CardDataFormat => {
  if (data.columns.length > 1) {
    return "formal";
  }
  return "simple";
};

const useCardState = ({ data, properties }: CardProps) => {
  const [calculatedProperties, setCalculatedProperties] = useState<CardState>({
    loading: true,
    label: null,
    value: null,
    type: null,
  });

  useEffect(() => {
    if (!data) {
      return;
    }

    if (
      !data.columns ||
      !data.rows ||
      data.columns.length === 0 ||
      data.rows.length === 0
    ) {
      setCalculatedProperties({
        loading: false,
        label: null,
        value: null,
        type: null,
      });
      return;
    }

    const dataFormat = getDataFormat(data);

    if (dataFormat === "simple") {
      const firstCol = data.columns[0];
      const row = data.rows[0];
      setCalculatedProperties({
        loading: false,
        label: firstCol.name,
        value: row[0],
        type: properties.type ? properties.type : null,
      });
    } else {
      const labelColIndex = getColumnIndex(data.columns, "label");
      const formalLabel =
        labelColIndex >= 0 ? get(data, `rows[0][${labelColIndex}]`) : null;
      const valueColIndex = getColumnIndex(data.columns, "value");
      const formalValue =
        valueColIndex >= 0 ? get(data, `rows[0][${valueColIndex}]`) : null;
      const typeColIndex = getColumnIndex(data.columns, "type");
      const formalType =
        typeColIndex >= 0 ? get(data, `rows[0][${typeColIndex}]`) : null;
      setCalculatedProperties({
        loading: false,
        label: formalLabel,
        value: formalValue,
        type: formalType,
      });
    }
  }, [data, properties]);

  return calculatedProperties;
};

const Card = (props: CardProps) => {
  const state = useCardState(props);

  // return (
  //   <div className="w-full h-24 bg-ok border border-black-scale-2 text-foreground">
  //     {state.value}
  //   </div>
  // );

  // return (
  //   <div className="px-4 py-5 bg-green-200 shadow rounded-lg overflow-hidden sm:p-6">
  //     <dt className="text-sm font-medium text-gray-500 truncate">
  //       {state.label}
  //     </dt>
  //     <dd className="mt-1 text-3xl font-semibold text-gray-900">
  //       {state.value}
  //     </dd>
  //   </div>
  // );

  return (
    <div
      className={
        "flex-col overflow-hidden shadow rounded-lg " +
        getWrapperClasses(state.type)
      }
    >
      <div className="flex-grow">
        <div className="flex">
          <div className="w-0 flex-1 px-4 py-5 sm:p-6">
            <dt
              className={
                "text-sm font-medium truncate " + getTextClasses(state.type)
              }
            >
              {state.loading && "Loading..."}
              {!state.loading && !state.label && (
                <NilIcon className="h-5 w-5 text-black-scale-4" />
              )}
              {!state.loading && state.label}
            </dt>
            <dd className="flex items-baseline mt-2">
              <div
                className={"text-3xl font-light " + getTextClasses(state.type)}
              >
                {state.loading && (
                  <LoadingIndicator className="h-8 w-8 text-black-scale-4" />
                )}
                {!state.loading &&
                  (state.value === null || state.value === undefined) && (
                    <NilIcon className="h-8 w-8 text-black-scale-4" />
                  )}
                <IntegerDisplay
                  className="md:hidden"
                  num={state.value}
                  startAt="k"
                />
                <IntegerDisplay
                  className="hidden md:inline"
                  num={state.value}
                  startAt="m"
                />
              </div>
            </dd>
          </div>
          {state.type === "alert" && (
            <div className="py-2 px-3">
              <AlertIcon className="text-white opacity-30 text-3xl h-8 w-8" />
            </div>
          )}
          {state.type === "ok" && (
            <div className="py-2 px-3">
              <OKIcon className="block text-white opacity-30 h-8 w-8" />
            </div>
          )}
          {state.type === "info" && (
            <div className="py-2 px-3">
              <InfoIcon className="text-white opacity-30 text-3xl h-8 w-8" />
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

const CardWrapper = (props: CardProps) => {
  if (get(props, "properties.type") === "table") {
    return <Table {...props} />;
  }
  return <Card {...props} />;
};

export default CardWrapper;

export { getTextClasses, getWrapperClasses };
