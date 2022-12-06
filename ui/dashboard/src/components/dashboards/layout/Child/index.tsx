import { DashboardLayoutNode, PanelDefinition } from "../../../../types";
import { getComponent } from "../../index";
import { getNodeAndEdgeDataFormat } from "../../common/useNodeAndEdgeData";
import { NodeAndEdgeProperties } from "../../common/types";

type ChildProps = {
  layoutDefinition: DashboardLayoutNode;
  panelDefinition: PanelDefinition;
  showPanelControls?: boolean;
};

const Child = ({
  layoutDefinition,
  panelDefinition,
  showPanelControls = true,
}: ChildProps) => {
  const Panel = getComponent("panel");
  switch (layoutDefinition.panel_type) {
    case "benchmark":
    case "control":
      const Benchmark = getComponent("benchmark");
      return (
        <Benchmark
          {...(layoutDefinition as PanelDefinition)}
          showControls={showPanelControls}
        />
      );
    case "card":
      const Card = getComponent("card");
      return (
        <Panel definition={panelDefinition} showControls={showPanelControls}>
          <Card {...panelDefinition} />
        </Panel>
      );
    case "chart":
      const Chart = getComponent("chart");
      return (
        <Panel
          definition={panelDefinition}
          ready={!!panelDefinition.data}
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
        <Panel definition={panelDefinition} showControls={showPanelControls}>
          <ErrorPanel {...panelDefinition} />
        </Panel>
      );
    case "flow": {
      const Flow = getComponent("flow");
      const format = getNodeAndEdgeDataFormat(
        panelDefinition.properties as NodeAndEdgeProperties
      );
      return (
        <Panel
          definition={panelDefinition}
          ready={format === "NODE_AND_EDGE" || !!panelDefinition.data}
          showControls={showPanelControls}
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
        panelDefinition.properties as NodeAndEdgeProperties
      );
      return (
        <Panel
          definition={panelDefinition}
          ready={format === "NODE_AND_EDGE" || !!panelDefinition.data}
          showControls={showPanelControls}
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
        panelDefinition.properties as NodeAndEdgeProperties
      );
      return (
        <Panel
          definition={panelDefinition}
          ready={format === "NODE_AND_EDGE" || !!panelDefinition.data}
          showControls={showPanelControls}
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
          ready={
            panelDefinition.sql
              ? !!panelDefinition.data
              : !!panelDefinition.properties?.src
          }
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
          showControls={
            showPanelControls &&
            (panelDefinition.title || panelDefinition.display_type === "table")
          }
        >
          <Input {...panelDefinition} />
        </Panel>
      );
    case "table":
      const Table = getComponent("table");
      return (
        <Panel
          definition={panelDefinition}
          ready={!!panelDefinition.data}
          showControls={showPanelControls}
        >
          <Table {...panelDefinition} />
        </Panel>
      );
    case "text":
      const Text = getComponent("text");
      return (
        <Panel definition={panelDefinition} showControls={false}>
          <Text {...panelDefinition} />
        </Panel>
      );
    default:
      return null;
  }
};

export default Child;
