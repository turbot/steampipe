import DashboardIcon, {
  useDashboardIconType,
} from "../../common/DashboardIcon";
import Icon from "../../../Icon";
import IntegerDisplay from "../../../IntegerDisplay";
import RowProperties, { RowPropertiesTitle } from "./RowProperties";
import Tooltip from "./Tooltip";
import useChartThemeColors from "../../../../hooks/useChartThemeColors";
import usePaginatedList from "../../../../hooks/usePaginatedList";
import {
  Category,
  CategoryFields,
  CategoryFold,
  FoldedNode,
  KeyValuePairs,
} from "../../common/types";
import { classNames } from "../../../../utils/styles";
import { ExpandedNodeInfo, useGraph } from "../common/useGraph";
import { getColorOverride } from "../../common";
import { Handle } from "reactflow";
import { memo, useEffect, useMemo, useState } from "react";
import { renderInterpolatedTemplates } from "../../../../utils/template";
import { ThemeNames } from "../../../../hooks/useTheme";
import { useDashboard } from "../../../../hooks/useDashboard";

type AssetNodeProps = {
  id: string;
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
  category: Category | undefined;
  foldedNodes: FoldedNode[] | undefined;
};

type FoldNodeIconProps = {
  collapseNodes: (foldedNodes: FoldedNode[]) => void;
  expandedNodeInfo: ExpandedNodeInfo | undefined;
};

type FoldedNodeLabelProps = {
  category: Category | undefined;
  fold: CategoryFold | undefined;
};

type FoldedNodeTooltipTitleProps = {
  category: Category | undefined;
  foldedNodesCount: number;
};

type FolderNodeTooltipNodesProps = {
  foldedNodes: FoldedNode[] | undefined;
};

const FoldedNodeTooltipTitle = ({
  category,
  foldedNodesCount,
}: FoldedNodeTooltipTitleProps) => {
  const themeColors = useChartThemeColors();
  return (
    <div className="flex flex-col space-y-1">
      {category && (
        <span
          className="block text-foreground-lighter text-xs"
          style={{ color: getColorOverride(category.color, themeColors) }}
        >
          {category.title || category.name}
        </span>
      )}
      <strong className="block">
        <IntegerDisplay num={foldedNodesCount} /> nodes
      </strong>
    </div>
  );
};

const FolderNodeTooltipNodes = ({
  foldedNodes,
}: FolderNodeTooltipNodesProps) => {
  const { visibleItems, hasMore, loadMore } = usePaginatedList(foldedNodes, 5);

  return (
    <div className="max-h-1/2-screen space-y-2">
      <div className="h-full overflow-y-auto">
        {(visibleItems || []).map((n) => (
          <div key={n.id}>{n.title || n.id}</div>
        ))}
        {hasMore && (
          <div
            className="flex items-center text-sm cursor-pointer space-x-1 text-link"
            onClick={loadMore}
          >
            <span>More</span>
            <Icon className="w-4 h-4" icon="arrow-long-down" />
          </div>
        )}
      </div>
    </div>
  );
};

const FoldedNodeCountBadge = ({
  category,
  foldedNodes,
}: FoldedNodeCountBadgeProps) => {
  const { expandNode } = useGraph();
  if (!foldedNodes) {
    return null;
  }
  return (
    <Tooltip
      overlay={<FolderNodeTooltipNodes foldedNodes={foldedNodes} />}
      title={
        <FoldedNodeTooltipTitle
          category={category}
          foldedNodesCount={foldedNodes.length}
        />
      }
    >
      <div
        className="absolute -right-[4%] -top-[4%] items-center bg-info text-white rounded-full px-1.5 text-sm font-medium cursor-pointer"
        onClick={() => expandNode(foldedNodes, category?.name as string)}
      >
        <IntegerDisplay num={foldedNodes?.length || null} />
      </div>
    </Tooltip>
  );
};

const FoldNodeIcon = ({
  collapseNodes,
  expandedNodeInfo,
}: FoldNodeIconProps) => {
  if (!expandedNodeInfo) {
    return null;
  }
  return (
    <div
      className="absolute -right-[4%] -top-[4%] items-center bg-foreground-lightest text-foreground-light rounded-full p-0.5 text-sm font-medium cursor-pointer"
      title="Collapse"
      onClick={() => collapseNodes(expandedNodeInfo.foldedNodes)}
    >
      <Icon className="w-4 h-4" icon="arrows-pointing-in" />
    </div>
  );
};

const FoldedNodeLabel = ({ category, fold }: FoldedNodeLabelProps) => (
  <div className="flex space-x-1 items-center">
    {fold?.title && (
      <span className="cursor-pointer truncate" title={fold?.title}>
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

const NodeControls = () => {
  return (
    <div className="invisible peer-hover:visible absolute -left-[4%] -bottom-[4%] items-center bg-black-scale-2 p-1 rounded-full cursor-grab"></div>
  );
};

const AssetNode = ({
  id,
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
  const { collapseNodes, expandNode, expandedNodes } = useGraph();
  const {
    themeContext: { theme },
  } = useDashboard();
  const {
    components: { ExternalLink },
  } = useDashboard();
  const iconType = useDashboardIconType(icon);

  const [renderedHref, setRenderedHref] = useState<string | null>(null);

  const isExpandedNode = useMemo(
    () => !!expandedNodes[id],
    [id, expandedNodes]
  );

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

  const nodeGrabHandle = (
    <div className="custom-drag-handle absolute -left-[4%] -bottom-[4%] items-center bg-black-scale-2 p-1 rounded-full cursor-grab">
      <Icon className="w-4 h-4" icon="cursor-arrow-ripple" />
    </div>
  );

  const node = (
    <div
      className={classNames(
        "relative p-3 rounded-full w-[50px] h-[50px] leading-[50px] my-0 mx-auto border"
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
      {isFolded && (
        <FoldedNodeCountBadge category={category} foldedNodes={foldedNodes} />
      )}
      {isExpandedNode && (
        <FoldNodeIcon
          collapseNodes={collapseNodes}
          expandedNodeInfo={expandedNodes[id]}
        />
      )}
      {/*{nodeGrabHandle}*/}
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
        "absolute flex space-x-1 items-center justify-center bottom-0 px-1 text-sm mt-1 bg-dashboard-panel text-foreground whitespace-nowrap min-w-[35px] max-w-[150px]"
      )}
      onClick={
        isFolded && foldedNodes
          ? () => expandNode(foldedNodes, category?.name as string)
          : undefined
      }
    >
      {!isFolded && renderedHref && (
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
      <div className="relative cursor-auto h-[72px]">
        <div className="peer flex flex-col group items-center">
          {nodeWithProperties}
          {nodeLabel}
        </div>
        <NodeControls />
      </div>
    </>
  );
};

export default memo(AssetNode);
