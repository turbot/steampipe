import ErrorPanel from "../../Error";
import merge from "lodash/merge";
import useChartThemeColors from "../../../../hooks/useChartThemeColors";
import useNodeAndEdgeData from "../../common/useNodeAndEdgeData";
import {
  buildNodesAndEdges,
  buildTreeDataInputs,
  LeafNodeData,
} from "../../common";
import { Chart } from "../../charts/Chart";
import { getHierarchyComponent } from "..";
import { HierarchyProperties, HierarchyProps, HierarchyType } from "../types";
import { NodesAndEdges } from "../../common/types";
import { registerComponent } from "../../index";
import { useDashboard } from "../../../../hooks/useDashboard";

const getCommonBaseOptions = () => ({
  animation: false,
  tooltip: {
    trigger: "item",
    triggerOn: "mousemove",
  },
});

const getCommonBaseOptionsForHierarchyType = (type: HierarchyType = "tree") => {
  switch (type) {
    default:
      return {};
  }
};

const getSeriesForHierarchyType = (
  type: HierarchyType = "tree",
  data: LeafNodeData | undefined,
  properties: HierarchyProperties | undefined,
  nodesAndEdges: NodesAndEdges,
  themeColors
) => {
  if (!data) {
    return {};
  }
  const series: any[] = [];
  const seriesLength = 1;
  for (let seriesIndex = 0; seriesIndex < seriesLength; seriesIndex++) {
    switch (type) {
      case "tree": {
        const { data: treeData } = buildTreeDataInputs(nodesAndEdges);
        series.push({
          type: "tree",
          data: treeData,
          top: "1%",
          left: "7%",
          bottom: "1%",
          right: "20%",
          symbolSize: 7,
          label: {
            color: themeColors.foreground,
            position: "left",
            verticalAlign: "middle",
            align: "right",
          },
          leaves: {
            label: {
              position: "right",
              verticalAlign: "middle",
              align: "left",
            },
          },
          emphasis: {
            focus: "descendant",
          },
          expandAndCollapse: false,
          animationDuration: 550,
          animationDurationUpdate: 750,
        });
      }
    }
  }

  return { series };
};

const getOptionOverridesForHierarchyType = (
  type: HierarchyType = "tree",
  properties: HierarchyProperties | undefined
) => {
  if (!properties) {
    return {};
  }

  return {};
};

const buildHierarchyOptions = (props: HierarchyProps, themeColors) => {
  const nodesAndEdges = buildNodesAndEdges(
    props.categories,
    props.data,
    props.properties,
    themeColors
  );

  return merge(
    getCommonBaseOptions(),
    getCommonBaseOptionsForHierarchyType(props.display_type),
    getSeriesForHierarchyType(
      props.display_type,
      props.data,
      props.properties,
      nodesAndEdges,
      themeColors
    ),
    getOptionOverridesForHierarchyType(props.display_type, props.properties)
  );
};

const HierarchyWrapper = (props: HierarchyProps) => {
  const themeColors = useChartThemeColors();
  const {
    themeContext: { wrapperRef },
  } = useDashboard();

  const nodeAndEdgeData = useNodeAndEdgeData(
    props.data,
    props.properties,
    props.status
  );

  if (!wrapperRef) {
    return null;
  }

  if (
    !nodeAndEdgeData ||
    !nodeAndEdgeData.data ||
    !nodeAndEdgeData.data.rows ||
    nodeAndEdgeData.data.rows.length === 0
  ) {
    return null;
  }

  return (
    <Chart
      options={buildHierarchyOptions(
        {
          ...props,
          categories: nodeAndEdgeData.categories,
          data: nodeAndEdgeData.data,
          properties: nodeAndEdgeData.properties,
        },
        themeColors
      )}
      type={props.display_type || "tree"}
    />
  );
};

const renderHierarchy = (definition: HierarchyProps) => {
  // We default to tree diagram if not specified
  const { display_type = "tree" } = definition;

  const hierarchy = getHierarchyComponent(display_type);

  if (!hierarchy) {
    return <ErrorPanel error={`Unknown hierarchy type ${display_type}`} />;
  }

  const Component = hierarchy.component;
  return <Component {...definition} />;
};

const RenderHierarchy = (props: HierarchyProps) => {
  return renderHierarchy(props);
};

registerComponent("hierarchy", RenderHierarchy);

export default HierarchyWrapper;
