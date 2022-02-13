import Icon from "../../Icon";
import IntegerDisplay from "../../IntegerDisplay";
import LoadingIndicator from "../LoadingIndicator";
import Table from "../Table";
import {
  BasePrimitiveProps,
  ExecutablePrimitiveProps,
  LeafNodeData,
} from "../common";
import { classNames } from "../../../utils/styles";
import { get, isNumber, isObject } from "lodash";
import { getColumnIndex } from "../../../utils/data";
import { ThemeNames, useTheme } from "../../../hooks/useTheme";
import { useEffect, useState } from "react";

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

const getIconClasses = (type) => {
  switch (type) {
    case "info":
    case "ok":
    case "alert":
      return "text-white opacity-40 text-3xl";
    default:
      return "";
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
      icon?: string;
    };
  };

type CardDataFormat = "simple" | "formal";

interface CardState {
  loading: boolean;
  label: string | null;
  value: any | null;
  type: CardType;
  icon: string | null;
}

const getDataFormat = (data: LeafNodeData): CardDataFormat => {
  if (data.columns.length > 1) {
    return "formal";
  }
  return "simple";
};

const getIconForType = (type) => {
  if (!type) {
    return null;
  }

  switch (type) {
    case "alert":
      return "heroicons-solid:x-circle";
    case "ok":
      return "heroicons-solid:check";
    case "info":
      return "heroicons-solid:information-circle";
    default:
      return null;
  }
};

const useCardState = ({ data, properties }: CardProps) => {
  const [calculatedProperties, setCalculatedProperties] = useState<CardState>({
    loading: true,
    label: null,
    value: null,
    type: properties.type || null,
    icon: getIconForType(properties.type) || properties.icon || null,
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
        type: properties.type || null,
        icon: getIconForType(properties.type) || properties.icon || null,
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
        type: properties.type || null,
        icon: getIconForType(properties.type) || properties.icon || null,
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
      const iconColIndex = getColumnIndex(data.columns, "icon");
      const formalIcon =
        iconColIndex >= 0 ? get(data, `rows[0][${iconColIndex}]`) : null;
      setCalculatedProperties({
        loading: false,
        label: formalLabel,
        value: formalValue,
        type: formalType || properties.type || null,
        icon: formalIcon || properties.icon || null,
      });
    }
  }, [data, properties]);

  return calculatedProperties;
};

const Label = ({ value }) => {
  if (!value) {
    return null;
  }

  if (isObject(value)) {
    return JSON.stringify(value);
  }

  return value;
};

const Card = (props: CardProps) => {
  const state = useCardState(props);
  const { theme } = useTheme();

  return (
    <div
      className={classNames(
        "relative pt-4 px-3 pb-4 sm:px-4 rounded-lg overflow-hidden",
        getWrapperClasses(state.type)
      )}
    >
      <dt>
        <div className="absolute">
          {state.icon && (
            <Icon
              className={classNames(getIconClasses(state.type), "h-8 w-8")}
              icon={state.icon}
            />
          )}
        </div>
        <p
          className={classNames(
            "text-sm font-medium truncate",
            state.type === "alert" ||
              state.type === "ok" ||
              state.type === "info"
              ? "ml-11"
              : "ml-2",
            getTextClasses(state.type)
          )}
          title={state.label || undefined}
        >
          {state.loading && "Loading..."}
          {!state.loading && !state.label && (
            <Icon className="h-5 w-5" icon="heroicons-solid:minus" />
          )}
          {!state.loading && state.label}
        </p>
      </dt>
      <dd
        className={classNames(
          "flex items-baseline",
          state.type === "alert" || state.type === "ok" || state.type === "info"
            ? "ml-11"
            : "ml-2"
        )}
        title={state.value || undefined}
      >
        <p
          className={classNames(
            "text-4xl mt-1 font-semibold text-left",
            getTextClasses(state.type)
          )}
        >
          {state.loading && (
            <LoadingIndicator
              className={classNames(
                "h-8 w-8 mt-2",
                theme.name === ThemeNames.STEAMPIPE_DEFAULT
                  ? "text-black-scale-4"
                  : null
              )}
            />
          )}
          {!state.loading &&
            (state.value === null || state.value === undefined) && (
              <Icon className="h-10 w-10 -ml-1" icon="heroicons-solid:minus" />
            )}
          {state.value !== null &&
            state.value !== undefined &&
            !isNumber(state.value) && <Label value={state.value} />}
          {isNumber(state.value) && (
            <>
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
            </>
          )}
        </p>
      </dd>
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
