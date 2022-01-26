import Children from "../common/Children";
import LayoutPanel from "../common/LayoutPanel";
import PanelDetail from "../PanelDetail";
import { ReportDefinition, useReport } from "../../../../hooks/useReport";

interface ReportProps {
  definition: ReportDefinition;
  withPadding?: boolean;
}

// TODO allow full-screen of a panel
const Report = ({ definition, withPadding = false }: ReportProps) => (
  <LayoutPanel definition={definition} withPadding={withPadding}>
    {
      <>
        {definition.title ? (
          <Children
            children={[
              {
                name: `${definition.name}.root.text.title`,
                node_type: "text",
                properties: {
                  type: "markdown",
                  value: `# ${definition.title}`,
                },
              },
            ]}
          />
        ) : null}
      </>
    }
    <Children children={definition.children} />
  </LayoutPanel>
);

const ReportWrapper = () => {
  const { report, selectedReport, selectedPanel } = useReport();

  if (!report || !selectedReport) {
    return null;
  }

  if (selectedPanel) {
    return <PanelDetail definition={selectedPanel} />;
  }

  return <Report definition={report} withPadding={true} />;
};

export default ReportWrapper;

export { Report };
