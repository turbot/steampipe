import DashboardIcon from "../common/DashboardIcon";
import get from "lodash/get";
import has from "lodash/has";
import IntegerDisplay from "../../IntegerDisplay";
import isNumber from "lodash/isNumber";
import isObject from "lodash/isObject";
import LoadingIndicator from "../LoadingIndicator";
import useDeepCompareEffect from "use-deep-compare-effect";
import useTemplateRender from "../../../hooks/useTemplateRender";
import {
  BasePrimitiveProps,
  ExecutablePrimitiveProps,
  LeafNodeData,
} from "../common";
import { classNames } from "../../../utils/styles";
import { DashboardDataModeLive } from "../../../types";
import { getComponent, registerComponent } from "../index";
import {
  getIconClasses,
  getIconForType,
  getTextClasses,
  getWrapperClasses,
} from "../../../utils/card";
import { isRelativeUrl } from "../../../utils/url";
import { ThemeNames } from "../../../hooks/useTheme";
import { useDashboard } from "../../../hooks/useDashboard";
import { useEffect, useState } from "react";

const Table = getComponent("table");

export type CardType = "alert" | "info" | "ok" | "table" | null;

export type CardProps = Omit<BasePrimitiveProps, "display_type"> &
  ExecutablePrimitiveProps & {
    display_type?: CardType;
    properties: {
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

const useCardState = ({ data, sql, display_type, properties }: CardProps) => {
  const [calculatedProperties, setCalculatedProperties] = useState<CardState>({
    loading: !!sql,
    label: properties.label || null,
    value: isNumber(properties.value)
      ? properties.value
      : properties.value || null,
    type: display_type || null,
    icon: getIconForType(display_type, properties.icon),
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
        value: isNumber(properties.value)
          ? properties.value
          : properties.value || null,
        type: display_type || null,
        icon: getIconForType(display_type, properties.icon),
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
        value: row[firstCol.name],
        type: display_type || null,
        icon: getIconForType(display_type, properties.icon),
        href: properties.href || null,
      });
    } else {
      const formalLabel = get(data, "rows[0].label", null);
      const formalValue = get(data, `rows[0].value`, null);
      const formalType = get(data, `rows[0].type`, null);
      const formalIcon = get(data, `rows[0].icon`, null);
      const formalHref = get(data, `rows[0].href`, null);
      setCalculatedProperties({
        loading: false,
        label: formalLabel,
        value: formalValue,
        type: formalType || display_type || null,
        icon: getIconForType(
          formalType || display_type,
          formalIcon || properties.icon
        ),
        href: formalHref || properties.href || null,
      });
    }
  }, [data, display_type, properties]);

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
  const {
    components: { ExternalLink },
    dataMode,
  } = useDashboard();
  const state = useCardState(props);
  const [renderError, setRenderError] = useState<string | null>(null);
  const [renderedHref, setRenderedHref] = useState<string | null>(
    state.href || null
  );
  const textClasses = getTextClasses(state.type);
  const {
    themeContext: { theme },
  } = useDashboard();
  const { ready: templateRenderReady, renderTemplates } = useTemplateRender();

  useEffect(() => {
    if ((state.loading || !state.href) && (renderError || renderedHref)) {
      setRenderError(null);
      setRenderedHref(null);
    }
  }, [state.loading, state.href, renderError, renderedHref]);

  useDeepCompareEffect(() => {
    if (!templateRenderReady || state.loading || !state.href) {
      return;
    }
    // const { label, loading, value, ...rest } = state;
    const renderData = { ...state };
    if (props.data && props.data.columns && props.data.rows) {
      const row = props.data.rows[0];
      props.data.columns.forEach((col) => {
        if (!has(renderData, col.name)) {
          renderData[col.name] = row[col.name];
        }
      });
    }

    const doRender = async () => {
      const renderedResults = await renderTemplates(
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
        // We only want to render the HREF if it's live, or it's snapshot and absolute
        const isRelative = isRelativeUrl(
          renderedResults[0].card.result as string
        );
        setRenderedHref(
          dataMode !== DashboardDataModeLive && isRelative
            ? null
            : (renderedResults[0].card.result as string)
        );
        setRenderError(null);
      } else if (renderedResults[0].card.error) {
        setRenderError(renderedResults[0].card.error as string);
        setRenderedHref(null);
      }
    };
    doRender();
  }, [renderTemplates, templateRenderReady, state, props.data]);

  const card = (
    <div
      className={classNames(
        "relative pt-4 px-3 pb-4 sm:px-4 rounded-md overflow-hidden",
        getWrapperClasses(state.type)
      )}
    >
      <dt>
        <div className="absolute">
          <DashboardIcon
            className={classNames(getIconClasses(state.type), "h-8 w-8")}
            icon={state.icon}
          />
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
            <DashboardIcon
              className="h-5 w-5"
              icon="materialsymbols-outline:remove"
            />
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
              <DashboardIcon
                className="h-10 w-10"
                icon="materialsymbols-outline:remove"
              />
            )}
          {state.value !== null &&
            state.value !== undefined &&
            !isNumber(state.value) && <Label value={state.value} />}
          {isNumber(state.value) && (
            <>
              <IntegerDisplay num={state.value} startAt="100k" />
            </>
          )}
        </p>
      </dd>
    </div>
  );

  if (dataMode === DashboardDataModeLive && renderedHref) {
    return (
      <ExternalLink className="" to={renderedHref}>
        {card}
      </ExternalLink>
    );
  }

  return card;
};

const CardWrapper = (props: CardProps) => {
  if (props.display_type === "table") {
    // @ts-ignore
    return <Table {...props} />;
  }

  return <Card {...props} />;
};

registerComponent("card", CardWrapper);

export default CardWrapper;
