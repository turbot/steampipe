import { ContainerDefinition, PanelDefinition } from "../../../../types";
import { getComponent } from "../../index";
import { NodeAndEdgeProperties } from "../../common/types";
import { nodeAndEdgeResourceHasData } from "../../common";
import { useDashboard } from "../../../../hooks/useDashboard";

interface ChildrenProps {
  children: ContainerDefinition[] | PanelDefinition[] | undefined;
  showPanelControls?: boolean;
}

const Children = ({
  children = [],
  showPanelControls = true,
}: ChildrenProps) => {
  const { panelsMap } = useDashboard();
  const Panel = getComponent("panel");
  return (
    <>
      {children.map((child) => {
        const definition = panelsMap[child.name];
        if (!definition) {
          return null;
        }
        switch (child.panel_type) {
          case "benchmark":
          case "control":
            const Benchmark = getComponent("benchmark");
            return (
              <Benchmark
                key={definition.name}
                {...(child as PanelDefinition)}
              />
            );
          case "benchmark_tree":
            const BenchmarkTree = getComponent("benchmark_tree");
            return <BenchmarkTree key={definition.name} {...child} />;
          case "card":
            const Card = getComponent("card");
            return (
              <Panel
                key={definition.name}
                definition={definition}
                showControls={showPanelControls}
              >
                <Card {...definition} />
              </Panel>
            );
          case "chart":
            const Chart = getComponent("chart");
            return (
              <Panel
                key={definition.name}
                definition={definition}
                ready={!!definition.data}
                showControls={showPanelControls}
              >
                <Chart {...definition} />
              </Panel>
            );
          case "container":
            const Container = getComponent("container");
            return (
              <Container
                key={definition.name}
                definition={child}
                expandDefinition={child}
                showChildPanelControls={child.allow_child_panel_expand}
              />
            );
          case "dashboard":
            const Dashboard = getComponent("dashboard");
            return (
              <Dashboard
                key={definition.name}
                definition={definition}
                isRoot={false}
              />
            );
          case "error":
            const ErrorPanel = getComponent("error");
            return (
              <Panel
                key={definition.name}
                definition={definition}
                showControls={showPanelControls}
              >
                <ErrorPanel {...definition} />
              </Panel>
            );
          case "flow":
            const Flow = getComponent("flow");
            return (
              <Panel
                key={definition.name}
                definition={definition}
                ready={nodeAndEdgeResourceHasData(
                  definition.data,
                  definition.properties as NodeAndEdgeProperties,
                  panelsMap
                )}
                showControls={showPanelControls}
              >
                <Flow {...definition} />
              </Panel>
            );
          case "graph":
            const Graph = getComponent("graph");
            return (
              <Panel
                key={definition.name}
                definition={definition}
                ready={nodeAndEdgeResourceHasData(
                  definition.data,
                  definition.properties as NodeAndEdgeProperties,
                  panelsMap
                )}
                showControls={showPanelControls}
              >
                <Graph {...definition} />
              </Panel>
            );
          case "hierarchy":
            const Hierarchy = getComponent("hierarchy");
            return (
              <Panel
                key={definition.name}
                definition={definition}
                ready={nodeAndEdgeResourceHasData(
                  definition.data,
                  definition.properties as NodeAndEdgeProperties,
                  panelsMap
                )}
                showControls={showPanelControls}
              >
                <Hierarchy {...definition} />
              </Panel>
            );
          case "image":
            const Image = getComponent("image");
            return (
              <Panel
                key={definition.name}
                definition={definition}
                ready={
                  definition.sql
                    ? !!definition.data
                    : !!definition.properties?.src
                }
                showControls={showPanelControls}
              >
                <Image {...definition} />
              </Panel>
            );
          case "input":
            const Input = getComponent("input");
            return (
              <Panel
                key={definition.name}
                definition={definition}
                showControls={
                  showPanelControls &&
                  (child.title || child.display_type === "table")
                }
              >
                <Input {...definition} />
              </Panel>
            );
          case "table":
            const Table = getComponent("table");
            return (
              <Panel
                key={definition.name}
                definition={definition}
                ready={!!definition.data}
                showControls={showPanelControls}
              >
                <Table {...definition} />
              </Panel>
            );
          case "text":
            const Text = getComponent("text");
            return (
              <Panel
                key={definition.name}
                definition={definition}
                showControls={false}
              >
                <Text {...definition} />
              </Panel>
            );
          default:
            return null;
        }
      })}
    </>
  );
};

export default Children;
