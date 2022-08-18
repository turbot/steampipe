import DashboardIcon from "../../common/DashboardIcon";
import RowProperties from "./RowProperties";
import Tooltip from "./Tooltip";
import { CategoryFields, KeyValuePairs } from "../../common/types";
import { classNames } from "../../../../utils/styles";
import { Handle } from "react-flow-renderer";
import { memo, useEffect, useState } from "react";
import { renderInterpolatedTemplates } from "../../../../utils/template";
import { ThemeNames, useTheme } from "../../../../hooks/useTheme";
import { useDashboard } from "../../../../hooks/useDashboard";

interface AssetNodeProps {
  data: {
    color?: string;
    fields?: CategoryFields;
    href?: string;
    icon?: string;
    label: string;
    row_data?: KeyValuePairs;
    namedColors;
  };
}

const AssetNode = ({
  data: { color, fields, href, icon, row_data, label, namedColors },
}: AssetNodeProps) => {
  const { theme } = useTheme();
  const {
    components: { ExternalLink },
  } = useDashboard();

  const [renderedHref, setRenderedHref] = useState<string | null>(null);

  useEffect(() => {
    if (!href) {
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
  }, [href, renderInterpolatedTemplates, row_data, setRenderedHref]);

  const node = (
    <div
      className={classNames(
        "p-3 rounded-full w-[50px] h-[50px] leading-[50px] my-0 mx-auto border cursor-grab"
        // color ? "opacity-10" : null
      )}
      style={{
        // backgroundColor: color,
        borderColor: color ? color : namedColors.blackScale3,
      }}
    >
      <DashboardIcon
        className={classNames(
          "max-w-full",
          theme.name === ThemeNames.STEAMPIPE_DARK ? "brightness-[1.75]" : null
        )}
        icon={icon}
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
        {row_data && row_data.properties && (
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
        {(!row_data || !row_data.properties) && node}
        <div className="text-center text-sm mt-1 bg-dashboard-panel text-foreground min-w-[35px]">
          {renderedHref && (
            <ExternalLink to={renderedHref}>{label}</ExternalLink>
          )}
          {!renderedHref && label}
        </div>
      </div>
    </>
  );
};

export default memo(AssetNode);
