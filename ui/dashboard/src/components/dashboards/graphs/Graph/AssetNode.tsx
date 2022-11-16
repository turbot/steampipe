import DashboardIcon, {
  useDashboardIconType,
} from "../../common/DashboardIcon";
import IntegerDisplay from "../../../IntegerDisplay";
import RowProperties, { RowPropertiesTitle } from "./RowProperties";
import Tooltip from "./Tooltip";
import {
  Category,
  CategoryFields,
  CategoryFold,
  FoldedNode,
  KeyValuePairs,
} from "../../common/types";
import { classNames } from "../../../../utils/styles";
import { Handle } from "reactflow";
import { memo, useEffect, useState } from "react";
import { renderInterpolatedTemplates } from "../../../../utils/template";
import { ThemeNames } from "../../../../hooks/useTheme";
import { useDashboard } from "../../../../hooks/useDashboard";
import { useGraph } from "../common/useGraph";

type AssetNodeProps = {
  data: {
    category?: Category;
    color?: string;
    fields?: CategoryFields;
    fold?: CategoryFold;
    href?: string;
    icon?: string;
    isFolded: boolean;
    foldedNodes?: FoldedNode[];
    label: string;
    row_data?: KeyValuePairs;
    themeColors;
  };
};

type FoldedNodeCountBadgeProps = {
  foldedNodes: FoldedNode[] | undefined;
};

type FoldedNodeLabelProps = {
  category: Category | undefined;
  fold: CategoryFold | undefined;
};

const FoldedNodeCountBadge = ({ foldedNodes }: FoldedNodeCountBadgeProps) => {
  if (!foldedNodes) {
    return null;
  }
  return (
    <Tooltip
      overlay={
        <div className="max-h-1/2-screen space-y-2">
          <div className="h-full overflow-y-auto">
            {(foldedNodes || []).map((n) => (
              <div key={n.id}>{n.title || n.id}</div>
            ))}
          </div>
        </div>
      }
      title={`${foldedNodes.length} nodes`}
    >
      <div className="absolute -right-[4%] -top-[4%] items-center bg-info text-white rounded-full px-1.5 text-sm font-medium cursor-pointer">
        <IntegerDisplay num={foldedNodes?.length || null} />
      </div>
    </Tooltip>
  );
};

const FoldedNodeLabel = ({ category, fold }: FoldedNodeLabelProps) => (
  <div className="flex space-x-1 items-center">
    {fold?.title && (
      <span className="text-link cursor-pointer truncate" title={fold?.title}>
        {fold?.title}
      </span>
    )}
    {!fold?.title && (
      <span
        className="text-link cursor-pointer truncate"
        title={category?.name}
      >
        {category?.name}
      </span>
    )}
  </div>
);

const AssetNode = ({
  data: {
    category,
    color,
    fields,
    fold,
    href,
    icon,
    isFolded,
    foldedNodes,
    row_data,
    label,
    themeColors,
  },
}: AssetNodeProps) => {
  const { expandNode } = useGraph();
  const {
    themeContext: { theme },
  } = useDashboard();
  const {
    components: { ExternalLink },
  } = useDashboard();
  const iconType = useDashboardIconType(icon);

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
  }, [isFolded, href, row_data, setRenderedHref]);

  const node = (
    <div
      className={classNames(
        "relative p-3 rounded-full w-[50px] h-[50px] leading-[50px] my-0 mx-auto border cursor-grab"
      )}
      style={{
        // backgroundColor: color,
        borderColor: color ? color : themeColors.blackScale3,
        // borderWidth: row_data && row_data.id === "i-0aa50f7044a950942" ? 3 : 1,
        color: isFolded ? (color ? color : themeColors.blackScale3) : undefined,
      }}
    >
      <DashboardIcon
        className={classNames(
          "max-w-full",
          iconType === "icon" && !color ? "text-foreground-lighter" : null,
          theme.name === ThemeNames.STEAMPIPE_DARK ? "brightness-[1.75]" : null
        )}
        style={{
          color: color ? color : undefined,
        }}
        icon={isFolded ? fold?.icon : icon}
      />
      {isFolded && <FoldedNodeCountBadge foldedNodes={foldedNodes} />}
    </div>
  );

  // const primaryNode =
  //   row_data && row_data.id === "i-0aa50f7044a950942" ? (
  //     <div
  //       className="relative p-0.5 rounded-full border"
  //       style={{
  //         borderColor: color ? color : themeColors.blackScale3,
  //       }}
  //     >
  //       {node}
  //     </div>
  //   ) : (
  //     node
  //   );

  const nodeWithProperties =
    row_data && row_data.properties && !isFolded ? (
      <Tooltip
        overlay={
          <>
            {row_data && row_data.properties && (
              <RowProperties
                fields={fields || null}
                properties={row_data.properties}
              />
            )}
          </>
        }
        title={<RowPropertiesTitle category={category} title={label} />}
      >
        {node}
        {/*<div className="cursor-pointer text-black-scale-5">*/}
        {/*  <Icon className="w-4 h-4" icon="queue-list" />*/}
        {/*</div>*/}
      </Tooltip>
    ) : (
      node
    );

  const nodeLabel = (
    <div
      className={classNames(
        renderedHref ? "text-link cursor-pointer" : null,
        "absolute flex space-x-1 items-center justify-center -bottom-[20px] px-1 text-sm mt-1 bg-dashboard-panel text-foreground whitespace-nowrap min-w-[35px] max-w-[150px]"
      )}
      onClick={
        isFolded && foldedNodes
          ? () => expandNode(foldedNodes, category?.name as string)
          : undefined
      }
    >
      {renderedHref && (
        <ExternalLink className="truncate" to={renderedHref}>
          {label}
        </ExternalLink>
      )}
      {!renderedHref && (
        <>
          {!isFolded && (
            <span className="truncate" title={label}>
              {label}
            </span>
          )}
          {isFolded && <FoldedNodeLabel category={category} fold={fold} />}
        </>
      )}
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
      {/*<div className="max-w-[50px]">{label}</div>*/}
      <div className="relative flex flex-col items-center cursor-auto">
        {nodeWithProperties}
        {nodeLabel}
      </div>
    </>
  );
};

export default memo(AssetNode);
