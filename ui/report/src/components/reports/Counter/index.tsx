import Icon from "../../Icon";
import IntegerDisplay from "../../IntegerDisplay";
import LoadingIndicator from "../LoadingIndicator";
import React, { useEffect, useState } from "react";
import Table from "../Table";
import { alertIcon, infoIcon, nilIcon, okIcon } from "../../../constants/icons";
import {
  BasePrimitiveProps,
  ExecutablePrimitiveProps,
  LeafNodeData,
} from "../common";
import { get } from "lodash";
import { hasColumn } from "../../../utils/data";

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
      value: null,
      style: null,
    });

  useEffect(() => {
    if (!data) {
      return;
    }

    if (
      !data.columns ||
      !data.items ||
      data.columns.length === 0 ||
      data.items.length === 0
    ) {
      setCalculatedProperties({
        loading: false,
        label: null,
        value: null,
        style: null,
      });
      return;
    }

    const dataFormat = getDataFormat(data);

    if (dataFormat === "simple") {
      const item = data.items[0];
      const kvps = Object.entries(item);
      const [label, value] = kvps[0];
      // const label = get(data, "items[0][0]");
      // const value = get(data, "[1][0]");
      setCalculatedProperties({
        loading: false,
        label,
        value,
        style: properties.style ? properties.style : null,
      });
    } else {
      const hasLabelCol = hasColumn(data.columns, "label");
      const formalLabel = hasLabelCol ? get(data, `items[0].label`) : null;
      const hasValueCol = hasColumn(data.columns, "value");
      const formalValue = hasValueCol ? get(data, `items[0].value`) : null;
      const hasStyleCol = hasColumn(data.columns, "style");
      const formalStyle = hasStyleCol ? get(data, `items[0].style`) : null;
      setCalculatedProperties({
        loading: false,
        label: formalLabel,
        value: formalValue,
        style: formalStyle,
      });
    }
  }, [data, properties]);

  return calculatedProperties;
};

const Counter = (props: CounterProps) => {
  const state = useCounterState(props);
  //
  // return (
  //   <div className="w-full h-24 bg-ok border border-black-scale-2 text-foreground">
  //     {state.value}
  //   </div>
  // );
  //
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
        getWrapperClasses(state.style)
      }
    >
      <div className="flex-grow">
        <div className="flex">
          <div className="w-0 flex-1 px-4 py-5 sm:p-6">
            <dt
              className={
                "text-sm font-medium truncate " + getTextClasses(state.style)
              }
            >
              {state.loading && "Loading..."}
              {!state.loading && !state.label && <Icon icon={nilIcon} />}
              {!state.loading && state.label}
            </dt>
            <dd className="flex items-baseline mt-2">
              <div
                className={"text-3xl font-light " + getTextClasses(state.style)}
              >
                {state.loading && <LoadingIndicator />}
                {!state.loading &&
                  (state.value === null || state.value === undefined) && (
                    <Icon icon={nilIcon} />
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
            <div className="text-white opacity-30 text-3xl py-2 px-3">
              <Icon icon={alertIcon} />
            </div>
          )}
          {state.style === "ok" && (
            <div className="text-white opacity-30 text-3xl py-2 px-3">
              <Icon icon={okIcon} />
            </div>
          )}
          {state.style === "info" && (
            <div className="text-white opacity-30 text-3xl py-2 px-3">
              <Icon icon={infoIcon} />
            </div>
          )}
        </div>
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
