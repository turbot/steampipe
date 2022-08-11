import Properties from "./Properties";
import Tooltip from "./Tooltip";
import { Handle } from "react-flow-renderer";
import { memo, useEffect, useState } from "react";
import { renderInterpolatedTemplates } from "../../../../utils/template";
import { LeafNodeDataRow } from "../../common";
import { useDashboard } from "../../../../hooks/useDashboard";

interface AssetNodeProps {
  data: {
    href?: string;
    icon?: string;
    label: string;
    row_data?: LeafNodeDataRow;
  };
}

const AssetNode = ({
  data: { href, icon, row_data, label },
}: AssetNodeProps) => {
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
  }, [href, renderInterpolatedTemplates]);

  const node = (
    <div className="py-2 px-2 rounded-full w-[35px] h-[35px] text-sm leading-[35px] my-0 mx-auto border border-divide cursor-grab">
      {icon && <img className="max-w-full" src={icon} />}
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
            overlay={<Properties properties={row_data.properties} />}
            title={label}
          >
            {node}
          </Tooltip>
        )}
        {(!row_data || !row_data.properties) && node}
        <div className="text-center text-[7px] mt-1 bg-dashboard-panel text-foreground min-w-[35px]">
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
