import {
  DashboardLayoutNode,
  DashboardPanelType,
  PanelDefinition,
} from "../../../../types";
import { getComponent } from "../../index";
import { getNodeAndEdgeDataFormat } from "../../common/useNodeAndEdgeData";
import { NodeAndEdgeProperties } from "../../common/types";
import { useDashboard } from "../../../../hooks/useDashboard";

type ChildProps = {
  layoutDefinition: DashboardLayoutNode;
  panelDefinition: PanelDefinition;
  parentType: DashboardPanelType;
  showPanelControls?: boolean;
};

const Child = ({
  layoutDefinition,
  panelDefinition,
  parentType,
  showPanelControls = true,
}: ChildProps) => {
  const { diff } = useDashboard();
  const diff_panel = diff ? diff.panelsMap[panelDefinition.name] : null;
  const Panel = getComponent("panel");
  switch (layoutDefinition.panel_type) {
    case "benchmark":
    case "control":
      const Benchmark = getComponent("benchmark");
      return (
        <Benchmark
          {...(layoutDefinition as PanelDefinition)}
          diff_panels={diff ? diff.panelsMap : null}
          showControls={showPanelControls}
        />
      );
    case "card":
      const Card = getComponent("card");
      return (
        <Panel
          definition={panelDefinition}
          parentType={parentType}
          showControls={showPanelControls}
          showPanelStatus={false}
        >
          <Card {...panelDefinition} diff_panel={diff_panel} />
        </Panel>
      );
    case "chart":
      const Chart = getComponent("chart");
      return (
        <Panel
          definition={panelDefinition}
          parentType={parentType}
          showControls={showPanelControls}
        >
          <Chart {...panelDefinition} />
        </Panel>
      );
    case "container":
      const Container = getComponent("container");
      return <Container layoutDefinition={layoutDefinition} />;
    case "dashboard":
      const Dashboard = getComponent("dashboard");
      return <Dashboard definition={panelDefinition} isRoot={false} />;
    case "error":
      const ErrorPanel = getComponent("error");
      return (
        <Panel
          definition={panelDefinition}
          parentType={parentType}
          showControls={showPanelControls}
        >
          <ErrorPanel {...panelDefinition} />
        </Panel>
      );
    case "flow": {
      const Flow = getComponent("flow");
      const format = getNodeAndEdgeDataFormat(
        panelDefinition.properties as NodeAndEdgeProperties,
      );
      return (
        <Panel
          definition={panelDefinition}
          parentType={parentType}
          showPanelStatus={
            format === "LEGACY" ||
            panelDefinition.status === "cancelled" ||
            panelDefinition.status === "error"
          }
          // Node and edge format will show error info on the panel information
          showPanelError={format === "LEGACY"}
        >
          <Flow {...panelDefinition} />
        </Panel>
      );
    }
    case "graph": {
      const Graph = getComponent("graph");
      const format = getNodeAndEdgeDataFormat(
        panelDefinition.properties as NodeAndEdgeProperties,
      );
      return (
        <Panel
          definition={panelDefinition}
          parentType={parentType}
          showControls={showPanelControls}
          showPanelStatus={
            format === "LEGACY" ||
            panelDefinition.status === "cancelled" ||
            panelDefinition.status === "error"
          }
          // Node and edge format will show error info on the panel information
          showPanelError={format === "LEGACY"}
        >
          <Graph {...panelDefinition} />
        </Panel>
      );
    }
    case "hierarchy": {
      const Hierarchy = getComponent("hierarchy");
      const format = getNodeAndEdgeDataFormat(
        panelDefinition.properties as NodeAndEdgeProperties,
      );
      return (
        <Panel
          definition={panelDefinition}
          parentType={parentType}
          showPanelStatus={
            format === "LEGACY" ||
            panelDefinition.status === "cancelled" ||
            panelDefinition.status === "error"
          }
          // Node and edge format will show error info on the panel information
          showPanelError={format === "LEGACY"}
        >
          <Hierarchy {...panelDefinition} />
        </Panel>
      );
    }
    case "image":
      const Image = getComponent("image");
      return (
        <Panel
          definition={panelDefinition}
          parentType={parentType}
          showControls={showPanelControls}
        >
          <Image {...panelDefinition} />
        </Panel>
      );
    case "input":
      const Input = getComponent("input");
      return (
        <Panel
          definition={panelDefinition}
          parentType={parentType}
          showControls={
            showPanelControls &&
            (panelDefinition.title || panelDefinition.display_type === "table")
          }
          showPanelStatus={false}
        >
          <Input {...panelDefinition} />
        </Panel>
      );
    case "table":
      const Table = getComponent("table");
      return (
        <Panel
          definition={panelDefinition}
          parentType={parentType}
          showControls={showPanelControls}
        >
          <Table {...panelDefinition} />
        </Panel>
      );
    case "text":
      const Text = getComponent("text");
      return (
        <Panel
          definition={panelDefinition}
          parentType={parentType}
          showControls={false}
        >
          <Text {...panelDefinition} />
        </Panel>
      );
    default:
      return null;
  }
};

export default Child;
