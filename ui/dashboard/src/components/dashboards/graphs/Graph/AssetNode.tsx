import DashboardIcon from "../../common/DashboardIcon";
import RowProperties from "./RowProperties";
import Tooltip from "./Tooltip";
import { CategoryFields, KeyValuePairs } from "../../common/types";
import { CategoryFold } from "../../common";
import { classNames } from "../../../../utils/styles";
import { Handle } from "react-flow-renderer";
import { memo, useEffect, useState } from "react";
import { renderInterpolatedTemplates } from "../../../../utils/template";
import { ThemeNames, useTheme } from "../../../../hooks/useTheme";
import { useDashboard } from "../../../../hooks/useDashboard";
import { useGraph } from "../common/useGraph";

interface AssetNodeProps {
  data: {
    category?: string;
    color?: string;
    fields?: CategoryFields;
    fold?: CategoryFold;
    href?: string;
    icon?: string;
    isFolded: boolean;
    label: string;
    row_data?: KeyValuePairs;
    namedColors;
  };
}

const AssetNode = ({
  data: {
    category,
    color,
    fields,
    fold,
    href,
    icon,
    isFolded,
    row_data,
    label,
    namedColors,
  },
}: AssetNodeProps) => {
  const { expandCategory } = useGraph();
  const { theme } = useTheme();
  const {
    components: { ExternalLink },
  } = useDashboard();

  const [renderedHref, setRenderedHref] = useState<string | null>(null);

  useEffect(() => {
    if (isFolded || !href) {
      setRenderedHref(null);
      return;
    }

    const doRender = async () => {
      const renderedResults = await renderInterpolatedTemplates(
        { graph_node: href as string },
        [row_data || {}]
      );
      const rowRenderResult = renderedResults[0];
      if (rowRenderResult.graph_node.result) {
        setRenderedHref(rowRenderResult.graph_node.result as string);
      }
    };

    doRender();
  }, [isFolded, href, renderInterpolatedTemplates, row_data, setRenderedHref]);

  const node = (
    <div
      className={classNames(
        "p-3 rounded-full w-[50px] h-[50px] leading-[50px] my-0 mx-auto border cursor-grab"
        // color ? "opacity-10" : null
      )}
      style={{
        // backgroundColor: color,
        borderColor: color ? color : namedColors.blackScale3,
        color: isFolded ? (color ? color : namedColors.blackScale3) : undefined,
      }}
    >
      <DashboardIcon
        className={classNames(
          "max-w-full",
          theme.name === ThemeNames.STEAMPIPE_DARK ? "brightness-[1.75]" : null
        )}
        icon={isFolded ? fold?.icon : icon}
      />
    </div>
  );

  // Notes:
  // * The Handle elements seem to be required to allow the connectors to work.
  return (
    <>
      {/*@ts-ignore*/}
      <Handle type="target" />
      {/*@ts-ignore*/}
      <Handle type="source" />
      <div className="flex flex-col items-center cursor-auto">
        {row_data && row_data.properties && !isFolded && (
          <Tooltip
            overlay={
              <RowProperties
                fields={fields || null}
                properties={row_data.properties}
              />
            }
            title={label}
          >
            {node}
          </Tooltip>
        )}
        {(isFolded || !row_data || !row_data.properties) && node}
        <div
          className={classNames(
            isFolded ? "text-link cursor-pointer" : null,
            "text-center text-sm mt-1 bg-dashboard-panel text-foreground min-w-[35px]"
          )}
          onClick={
            isFolded ? () => expandCategory(category as string) : undefined
          }
        >
          {renderedHref && (
            <ExternalLink to={renderedHref}>{label}</ExternalLink>
          )}
          {!renderedHref && !isFolded && label}
          {!renderedHref && isFolded && fold?.title}
        </div>
      </div>
    </>
  );
};

export default memo(AssetNode);
