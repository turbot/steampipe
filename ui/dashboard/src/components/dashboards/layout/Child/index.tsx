import {
  DashboardLayoutNode,
  PanelDefinition,
  PanelsMap,
} from "../../../../types";
import { getComponent } from "../../index";
import { NodeAndEdgeProperties } from "../../common/types";
import { nodeAndEdgeResourceHasData } from "../../common";

type ChildProps = {
  layoutDefinition: DashboardLayoutNode;
  panelDefinition: PanelDefinition;
  panelsMap: PanelsMap;
  showPanelControls?: boolean;
};

const Child = ({
  layoutDefinition,
  panelDefinition,
  panelsMap,
  showPanelControls = true,
}: ChildProps) => {
  const Panel = getComponent("panel");
  switch (layoutDefinition.panel_type) {
    case "benchmark":
      const Benchmark = getComponent("benchmark");
      return <Benchmark {...(layoutDefinition as PanelDefinition)} />;
    case "benchmark_tree":
      const BenchmarkTree = getComponent("benchmark_tree");
      return <BenchmarkTree {...layoutDefinition} />;
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
      return (
        <Container
          definition={layoutDefinition}
          expandDefinition={layoutDefinition}
        />
      );
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
    case "flow":
      const Flow = getComponent("flow");
      return (
        <Panel
          definition={panelDefinition}
          ready={nodeAndEdgeResourceHasData(
            panelDefinition.data,
            panelDefinition.properties as NodeAndEdgeProperties,
            panelsMap
          )}
          showControls={showPanelControls}
        >
          <Flow {...panelDefinition} />
        </Panel>
      );
    case "graph":
      const Graph = getComponent("graph");
      return (
        <Panel
          definition={panelDefinition}
          ready={nodeAndEdgeResourceHasData(
            panelDefinition.data,
            panelDefinition.properties as NodeAndEdgeProperties,
            panelsMap
          )}
          showControls={showPanelControls}
        >
          <Graph {...panelDefinition} />
        </Panel>
      );
    case "hierarchy":
      const Hierarchy = getComponent("hierarchy");
      return (
        <Panel
          definition={panelDefinition}
          ready={nodeAndEdgeResourceHasData(
            panelDefinition.data,
            panelDefinition.properties as NodeAndEdgeProperties,
            panelsMap
          )}
          showControls={showPanelControls}
        >
          <Hierarchy {...panelDefinition} />
        </Panel>
      );
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
