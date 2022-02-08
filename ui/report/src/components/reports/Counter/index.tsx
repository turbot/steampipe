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
import { Link } from "react-router-dom";

const getWrapperClasses = (style) => {
  switch (style) {
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

const getTextClasses = (style) => {
  switch (style) {
    case "alert":
      return "text-alert-inverse";
    case "info":
      return "text-info-inverse";
    case "ok":
      return "text-ok-inverse";
    default:
      return "text-counter-inverse";
  }
};

export type CounterStyle = "alert" | "info" | "ok" | null;

export type CounterProps = BasePrimitiveProps &
  ExecutablePrimitiveProps & {
    properties: {
      type?: "table";
      value?: string;
      style?: CounterStyle;
    };
  };

type CounterDataFormat = "simple" | "formal";

interface CounterState {
  loading: boolean;
  label: string | null;
  link_text: string | null;
  link_url: string | null;
  value: number | null;
  style: CounterStyle;
}

const getDataFormat = (data: LeafNodeData): CounterDataFormat => {
  if (data.columns.length > 1) {
    return "formal";
  }
  return "simple";
};

const useCounterState = ({ data, properties }: CounterProps) => {
  const [calculatedProperties, setCalculatedProperties] =
    useState<CounterState>({
      loading: true,
      label: null,
      link_text: null,
      link_url: null,
      value: null,
      style: null,
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
        link_text: null,
        link_url: null,
        value: null,
        style: null,
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
        link_text: null,
        link_url: null,
        value: row[0],
        style: properties.style ? properties.style : null,
      });
    } else {
      const labelColIndex = getColumnIndex(data.columns, "label");
      const formalLabel =
        labelColIndex >= 0 ? get(data, `rows[0][${labelColIndex}]`) : null;
      const valueColIndex = getColumnIndex(data.columns, "value");
      const formalValue =
        valueColIndex >= 0 ? get(data, `rows[0][${valueColIndex}]`) : null;
      const styleColIndex = getColumnIndex(data.columns, "style");
      const formalStyle =
        styleColIndex >= 0 ? get(data, `rows[0][${styleColIndex}]`) : null;
      const linkTextColIndex = getColumnIndex(data.columns, "link_text");
      const formalLinkText =
        linkTextColIndex >= 0
          ? get(data, `rows[0][${linkTextColIndex}]`)
          : null;
      const linkUrlColIndex = getColumnIndex(data.columns, "link_url");
      const formalLinkUrl =
        linkUrlColIndex >= 0 ? get(data, `rows[0][${linkUrlColIndex}]`) : null;
      setCalculatedProperties({
        loading: false,
        label: formalLabel,
        link_text: formalLinkText,
        link_url: formalLinkUrl,
        value: formalValue,
        style: formalStyle,
      });
    }
  }, [data, properties]);

  return calculatedProperties;
};

const Counter = (props: CounterProps) => {
  const state = useCounterState(props);

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
        "flex-col overflow-hidden rounded-lg " + getWrapperClasses(state.style)
      }
    >
      <div className="flex-grow flex-col">
        <div className="flex">
          <div className="w-0 flex-1 px-4 py-5 sm:p-6">
            <dt
              className={
                "text-sm font-medium truncate " + getTextClasses(state.style)
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
                className={"text-3xl font-light " + getTextClasses(state.style)}
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
          {state.style === "alert" && (
            <div className="py-2 px-3">
              <AlertIcon className="text-white opacity-30 text-3xl h-8 w-8" />
            </div>
          )}
          {state.style === "ok" && (
            <div className="py-2 px-3">
              <OKIcon className="block text-white opacity-30 h-8 w-8" />
            </div>
          )}
          {state.style === "info" && (
            <div className="py-2 px-3">
              <InfoIcon className="text-white opacity-30 text-3xl h-8 w-8" />
            </div>
          )}
        </div>
        {state.link_url && (
          <div className="bg-black-scale-1 px-4 py-4 sm:px-6">
            <div className="text-sm">
              <Link to={state.link_url} className="font-medium text-link">
                {state.link_text || state.link_url}
              </Link>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

const CounterWrapper = (props: CounterProps) => {
  if (get(props, "properties.type") === "table") {
    return <Table {...props} />;
  }
  return <Counter {...props} />;
};

export default CounterWrapper;

export { getTextClasses, getWrapperClasses };
