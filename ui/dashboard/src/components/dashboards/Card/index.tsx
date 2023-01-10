import DashboardIcon from "../common/DashboardIcon";
import get from "lodash/get";
import has from "lodash/has";
import IntegerDisplay from "../../IntegerDisplay";
import isNumber from "lodash/isNumber";
import isObject from "lodash/isObject";
import LoadingIndicator from "../LoadingIndicator";
import useDeepCompareEffect from "use-deep-compare-effect";
import usePanelDependenciesStatus from "../../../hooks/usePanelDependenciesStatus";
import useTemplateRender from "../../../hooks/useTemplateRender";
import {
  BasePrimitiveProps,
  ExecutablePrimitiveProps,
  LeafNodeData,
} from "../common";
import { classNames } from "../../../utils/styles";
import {
  DashboardDataModeLive,
  DashboardRunState,
  PanelProperties,
} from "../../../types";
import { getComponent, registerComponent } from "../index";
import {
  getIconClasses,
  getIconForType,
  getTextClasses,
  getWrapperClasses,
} from "../../../utils/card";
import { HashLink } from "react-router-hash-link";
import { InputProperties } from "../inputs/types";
import { isRelativeUrl } from "../../../utils/url";
import { Location } from "react-router-dom";
import { PanelDependencyStatuses } from "../common/types";
import { ReactNode, useEffect, useState } from "react";
import { ThemeNames } from "../../../hooks/useTheme";
import { useDashboard } from "../../../hooks/useDashboard";
import { useLocation } from "react-router-dom";

const Table = getComponent("table");

export type CardType = "alert" | "info" | "ok" | "table" | null;

export type CardProps = PanelProperties &
  Omit<BasePrimitiveProps, "display_type"> &
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

type CardState = {
  loading: boolean;
  title: string | undefined;
  label: ReactNode;
  value: any | null;
  type: CardType;
  icon: string | null;
  href: string | null;
};

const getDataFormat = (data: LeafNodeData): CardDataFormat => {
  if (data.columns.length > 1) {
    return "formal";
  }
  return "simple";
};

const getCardRunningStatus = (
  panelDependenciesStatus: PanelDependencyStatuses,
  status: DashboardRunState,
  location: Location,
  label: string | undefined
) => {
  let title: string | undefined = undefined;
  let component: ReactNode = null;
  let finalStatus: DashboardRunState | null = null;
  if (status === "initialized") {
    title = "Initialized";
    component = title;
    finalStatus = "initialized";
  } else if (status === "blocked") {
    if (panelDependenciesStatus.inputsAwaitingValue.length > 0) {
      const firstInput = panelDependenciesStatus.inputsAwaitingValue[0];
      const inputTitle =
        firstInput.title ||
        (firstInput.properties as InputProperties).unqualified_name;
      title = `Awaiting input value: ${inputTitle}`;
      component = (
        <>
          Awaiting input value:{" "}
          <HashLink
            className="text-link"
            to={`${location.pathname}${
              location.search ? location.search : ""
            }#${firstInput.name}`}
          >
            {inputTitle}
          </HashLink>
        </>
      );
      finalStatus = "blocked";
    }

    if (
      panelDependenciesStatus.inputsAwaitingValue.length === 0 &&
      panelDependenciesStatus.total === 0
    ) {
      title = "Running...";
      component = title;
      finalStatus = "running";
    }

    if (
      panelDependenciesStatus.inputsAwaitingValue.length === 0 &&
      panelDependenciesStatus.total > 0
    ) {
      title = `Running ${
        panelDependenciesStatus.status.complete.total +
        panelDependenciesStatus.status.running.total +
        1
      } of ${panelDependenciesStatus.total + 1}...`;
      component = title;
      finalStatus = "running";
    }
  } else if (status === "running") {
    if (
      panelDependenciesStatus.inputsAwaitingValue.length === 0 &&
      panelDependenciesStatus.total === 0
    ) {
      title = "Running...";
      component = title;
      finalStatus = "running";
    }

    if (
      panelDependenciesStatus.inputsAwaitingValue.length === 0 &&
      panelDependenciesStatus.total > 0
    ) {
      title = `Running ${
        panelDependenciesStatus.status.complete.total +
        panelDependenciesStatus.status.running.total +
        1
      } of ${panelDependenciesStatus.total + 1}...`;
      component = title;
      finalStatus = "running";
    }
  } else if (status === "cancelled") {
    title = "Cancelled";
    component = title;
    finalStatus = "cancelled";
  } else if (status === "error") {
    title = "Error";
    component = title;
    finalStatus = "error";
  } else if (status === "complete") {
    title = label;
    component = title;
    finalStatus = "complete";
  }

  return {
    title,
    component,
    status: finalStatus,
  };
};

const useCardState = ({
  data,
  display_type,
  properties,
  status,
}: CardProps) => {
  const location = useLocation();
  const panelDependenciesStatus = usePanelDependenciesStatus();

  const [calculatedProperties, setCalculatedProperties] = useState<CardState>(
    () => {
      const runningStatus = getCardRunningStatus(
        panelDependenciesStatus,
        status,
        location,
        properties.label
      );
      return {
        loading: runningStatus.status === "running",
        title: runningStatus.title,
        label: runningStatus.component,
        value: isNumber(properties.value)
          ? properties.value
          : properties.value || null,
        type: display_type || null,
        icon: getIconForType(
          display_type,
          status === "initialized"
            ? "materialsymbols-solid:start"
            : status === "blocked"
            ? "materialsymbols-solid:block"
            : properties.icon
        ),
        href: properties.href || null,
      };
    }
  );

  useEffect(() => {
    if (
      !data ||
      !data.columns ||
      !data.rows ||
      data.columns.length === 0 ||
      data.rows.length === 0
    ) {
      const runningStatus = getCardRunningStatus(
        panelDependenciesStatus,
        status,
        location,
        properties.label
      );
      setCalculatedProperties({
        loading: runningStatus.status === "running",
        title: runningStatus.title,
        label: runningStatus.component,
        value: isNumber(properties.value)
          ? properties.value
          : properties.value || null,
        type: display_type || null,
        icon: getIconForType(
          display_type,
          status === "initialized"
            ? "materialsymbols-solid:start"
            : status === "blocked"
            ? "materialsymbols-solid:block"
            : properties.icon
        ),
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
        title: firstCol.name,
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
        title: formalLabel,
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
  }, [
    data,
    display_type,
    location,
    panelDependenciesStatus,
    properties,
    status,
  ]);

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
          title={state.title}
        >
          {state.label}
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
