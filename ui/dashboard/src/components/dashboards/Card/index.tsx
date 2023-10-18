import DashboardIcon from "../common/DashboardIcon";
import has from "lodash/has";
import IntegerDisplay from "../../IntegerDisplay";
import isNumber from "lodash/isNumber";
import isObject from "lodash/isObject";
import LoadingIndicator from "../LoadingIndicator";
import useDeepCompareEffect from "use-deep-compare-effect";
import useTemplateRender from "../../../hooks/useTemplateRender";
import { BasePrimitiveProps, ExecutablePrimitiveProps } from "../common";
import {
  CardDataProcessor,
  CardDiffState,
  CardType,
} from "../data/CardDataProcessor";
import { classNames } from "../../../utils/styles";
import { PanelDefinition, PanelProperties } from "../../../types";
import { getComponent, registerComponent } from "../index";
import {
  getIconClasses,
  getTextClasses,
  getWrapperClasses,
} from "../../../utils/card";
import { IDiffProperties } from "../data/types";
import { ThemeNames } from "../../../hooks/useTheme";
import { useDashboard } from "../../../hooks/useDashboard";
import { useEffect, useState } from "react";

const Table = getComponent("table");

export interface CardProperties extends IDiffProperties {
  label?: string;
  value?: any;
  icon?: string;
  href?: string;
}

export type CardProps = PanelProperties &
  Omit<BasePrimitiveProps, "display_type"> &
  ExecutablePrimitiveProps & {
    display_type?: CardType;
    properties: CardProperties;
  } & {
    diff_panel: PanelDefinition;
  };

type CardState = {
  loading: boolean;
  label: string | null;
  value: any | null;
  type: CardType;
  icon: string | null;
  href: string | null;
  diff?: CardDiffState;
};

interface CardDiffDisplayProps {
  diff: CardDiffState | undefined;
  type: CardType;
}

// TODO diffing
// Need to know we're in diff mode
// Need data to diff against
// Need to be able to diff said data against current data
// Need to try to infer state of the change as best as possible
// e.g. a card going from alarm 10 to alarm 5 is good, so it's down 100% / green
//      a card going from alarm 10 to alarm 20 is bad, so it's up 100% / red
//      a card going from alarm 10 to alarm 10 is neutral, so it's no change
//      a card going from alarm 10 to ok 10 is good, so it's no change in value, but change in state

const useCardState = ({
  data,
  diff_panel,
  display_type,
  properties,
  status,
}: CardProps) => {
  const [calculatedProperties, setCalculatedProperties] = useState<CardState>(
    new CardDataProcessor().getDefaultState(status, properties, display_type),
  );

  useEffect(() => {
    const diff = new CardDataProcessor();
    setCalculatedProperties(
      diff.buildCardState(data, diff_panel, display_type, properties, status),
    );
  }, [data, diff_panel, display_type, properties, setCalculatedProperties, status]);

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

const CardDiffDisplay = ({ diff, type }: CardDiffDisplayProps) => {
  if (!diff || diff.direction === "none") {
    return null;
  }
  return (
    <div
      className={classNames(
        "inline-flex rounded-lg px-2 font-medium md:mt-2 lg:mt-0 space-x-1 text-lg",
        diff.status === "ok" ? "text-ok" : null,
        diff.status === "alert" ? "text-alert" : null,
        diff.status === "alert" ? "text-severity" : null,
      )}
    >
      <span aria-hidden="true" className={classNames("self-end")}>
        {diff.direction === "up" ? "↑" : diff.direction === "down" ? "↓" : null}
      </span>
      <span className="sr-only">
        {" "}
        {diff.direction === "up" ? "Increased" : "Decreased"} by{" "}
      </span>
      {(diff.direction === "up" || diff.direction === "down") && (
        <>
          {/*@ts-ignore*/}
          <IntegerDisplay num={diff.value || null} />
        </>
      )}
    </div>
  );
};

const Card = (props: CardProps) => {
  const ExternalLink = getComponent("external_link");
  const state = useCardState(props);
  const [renderError, setRenderError] = useState<string | null>(null);
  const [renderedHref, setRenderedHref] = useState<string | null>(
    state.href || null,
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
        [renderData],
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
  }, [renderTemplates, templateRenderReady, state, props.data]);

  const card = (
    <div
      className={classNames(
        "relative pt-4 px-3 pb-4 sm:px-4 rounded-md overflow-hidden",
        getWrapperClasses(state.type),
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
            "text-sm font-semibold truncate text-foreground",
            state.icon ? "ml-11" : "ml-2",
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
          "flex items-baseline space-x-4",
          state.icon ? "ml-11" : "ml-2",
        )}
        title={state.value || undefined}
      >
        <p
          className={classNames(
            "text-4xl mt-1 font-semibold text-left truncate",
            textClasses,
          )}
        >
          {state.loading && (
            <LoadingIndicator
              className={classNames(
                "h-9 w-9 mt-1",
                theme.name === ThemeNames.STEAMPIPE_DEFAULT
                  ? "text-black-scale-4"
                  : null,
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
        <CardDiffDisplay diff={state.diff} type={state.type} />
      </dd>
    </div>
  );

  if (renderedHref) {
    return <ExternalLink to={renderedHref}>{card}</ExternalLink>;
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
