import ExternalLink from "../../ExternalLink";
import Icon from "../../Icon";
import IntegerDisplay from "../../IntegerDisplay";
import LoadingIndicator from "../LoadingIndicator";
import Table from "../Table";
import useDeepCompareEffect from "use-deep-compare-effect";
import {
  BasePrimitiveProps,
  ExecutablePrimitiveProps,
  LeafNodeData,
} from "../common";
import { classNames } from "../../../utils/styles";
import { get, has, isNumber, isObject } from "lodash";
import { getColumnIndex } from "../../../utils/data";
import { renderInterpolatedTemplates } from "../../../utils/template";
import { ThemeNames } from "../../../hooks/useTheme";
import { useDashboard } from "../../../hooks/useDashboard";
import { useEffect, useState } from "react";
import { usePanel } from "../../../hooks/usePanel";

const getWrapperClasses = (type) => {
  switch (type) {
    case "alert":
      return "bg-alert";
    case "info":
      return "bg-info";
    case "ok":
      return "bg-ok";
    default:
      return "bg-dashboard-panel shadow-sm";
  }
};

const getIconClasses = (type) => {
  switch (type) {
    case "info":
    case "ok":
    case "alert":
      return "text-white opacity-40 text-3xl";
    default:
      return "text-black-scale-4 text-3xl";
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

type CardType = "alert" | "info" | "ok" | "table" | null;

export type CardProps = BasePrimitiveProps &
  ExecutablePrimitiveProps & {
    properties: {
      type?: CardType;
      label?: string;
      value?: string;
      icon?: string;
      href?: string;
    };
  };

type CardDataFormat = "simple" | "formal";

interface CardState {
  loading: boolean;
  label: string | null;
  value: any | null;
  type: CardType;
  icon: string | null;
  href: string | null;
}

const getDataFormat = (data: LeafNodeData): CardDataFormat => {
  if (data.columns.length > 1) {
    return "formal";
  }
  return "simple";
};

const getIconForType = (type, icon) => {
  if (!type && !icon) {
    return null;
  }

  if (icon) {
    return icon;
  }

  switch (type) {
    case "alert":
      return "heroicons-solid:exclamation-circle";
    case "ok":
      return "heroicons-solid:check-circle";
    case "info":
      return "heroicons-solid:information-circle";
    default:
      return null;
  }
};

const useCardState = ({ data, sql, properties }: CardProps) => {
  const [calculatedProperties, setCalculatedProperties] = useState<CardState>({
    loading: !!sql,
    label: properties.label || null,
    value: properties.value || null,
    type: properties.type || null,
    icon: getIconForType(properties.type, properties.icon),
    href: properties.href || null,
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
        label: properties.label || null,
        value: properties.value || null,
        type: properties.type || null,
        icon: getIconForType(properties.type, properties.icon),
        href: properties.href || null,
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
        icon: getIconForType(properties.type, properties.icon),
        href: properties.href || null,
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
      const hrefColIndex = getColumnIndex(data.columns, "href");
      const formalHref =
        hrefColIndex >= 0 ? get(data, `rows[0][${hrefColIndex}]`) : null;
      setCalculatedProperties({
        loading: false,
        label: formalLabel,
        value: formalValue,
        type: formalType || properties.type || null,
        icon: getIconForType(
          formalType || properties.type,
          formalIcon || properties.icon
        ),
        href: formalHref || properties.href || null,
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
  const [renderedHref, setRenderedHref] = useState<string | null>(
    state.href || null
  );
  const [, setRenderError] = useState<string | null>(null);
  const textClasses = getTextClasses(state.type);
  const { setZoomIconClassName } = usePanel();
  const {
    themeContext: { theme },
  } = useDashboard();

  useEffect(() => {
    setZoomIconClassName(textClasses ? textClasses : "");
  }, [setZoomIconClassName, textClasses]);

  useDeepCompareEffect(() => {
    if (state.loading || !state.href) {
      setRenderedHref(null);
      setRenderError(null);
      return;
    }
    // const { label, loading, value, ...rest } = state;
    const renderData = { ...state };
    if (props.data && props.data.columns && props.data.rows) {
      const row = props.data.rows[0];
      props.data.columns.forEach((col, index) => {
        if (!has(renderData, col.name)) {
          renderData[col.name] = row[index];
        }
      });
    }

    const doRender = async () => {
      const renderedResults = await renderInterpolatedTemplates(
        { card: state.href as string },
        [renderData]
      );
      if (
        !renderedResults ||
        renderedResults.length === 0 ||
        !renderedResults[0].card
      ) {
        setRenderedHref(null);
        setRenderError(null);
      } else if (renderedResults[0].card.result) {
        setRenderedHref(renderedResults[0].card.result as string);
        setRenderError(null);
      } else if (renderedResults[0].card.error) {
        setRenderError(renderedResults[0].card.error as string);
        setRenderedHref(null);
      }
    };
    doRender();
  }, [state, props.data]);

  const card = (
    <div
      className={classNames(
        "relative pt-4 px-3 pb-4 sm:px-4 m-0.5 rounded-md overflow-hidden",
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
            state.icon ? "ml-11" : "ml-2",
            textClasses
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
          state.icon ? "ml-11" : "ml-2"
        )}
        title={state.value || undefined}
      >
        <p
          className={classNames(
            "text-4xl mt-1 font-semibold text-left truncate",
            textClasses
          )}
        >
          {state.loading && (
            <LoadingIndicator
              className={classNames(
                "h-9 w-9 mt-1",
                theme.name === ThemeNames.STEAMPIPE_DEFAULT
                  ? "text-black-scale-4"
                  : null
              )}
            />
          )}
          {!state.loading &&
            (state.value === null || state.value === undefined) && (
              <Icon className="h-10 w-10" icon="heroicons-solid:minus" />
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

  if (renderedHref) {
    return (
      <ExternalLink className="" to={renderedHref}>
        {card}
      </ExternalLink>
    );
  }

  return card;
};

const CardWrapper = (props: CardProps) => {
  if (get(props, "properties.type") === "table") {
    // @ts-ignore
    return <Table {...props} />;
  }

  return <Card {...props} />;
};

export default CardWrapper;

export { getTextClasses, getWrapperClasses };
