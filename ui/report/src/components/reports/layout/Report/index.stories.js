import Report from "./index";
import { ReportContext } from "../../../../hooks/useReport";

const story = {
  title: "Layout/Report",
  component: Report,
};

export default story;

const Template = (args) => (
  <ReportContext.Provider value={{ dispatch: () => {}, report: args.report }}>
    <Report />
  </ReportContext.Provider>
);

export const Basic = Template.bind({});
Basic.args = {
  report: {
    name: "report.basic",
    title: "Basic Report",
    children: [
      {
        name: "text.header",
        type: "markdown",
        value: "## Basic Report",
      },
    ],
  },
};

export const TwoColumn = Template.bind({});
TwoColumn.args = {
  report: {
    name: "report.two_column",
    title: "Two Column Report",
    children: [
      {
        name: "text.header_1",
        type: "markdown",
        value: "## Column 1",
        width: 6,
      },
      {
        name: "text.header_2",
        type: "markdown",
        value: "## Column 2",
        width: 6,
      },
    ],
  },
};

export const LayoutContainer = Template.bind({});
LayoutContainer.args = {
  report: {
    name: "report.layout_container",
    title: "Layout Container Report",
    children: [
      {
        name: "container.wrapper",
        children: [
          {
            name: "text.title",
            type: "markdown",
            value: "## IAM Report",
          },
          {
            name: "chart.barchart",
            type: "bar",
            data: [
              ["Type", "Count"],
              ["User", 12],
              ["Policy", 93],
              ["Role", 48],
            ],
            title: "AWS IAM Entities",
          },
        ],
      },
    ],
  },
};

export const TwoColumnContainerLayout = Template.bind({});
TwoColumnContainerLayout.args = {
  report: {
    name: "report.layout_container",
    title: "Layout Container Report",
    children: [
      {
        name: "container.left",
        width: 6,
        children: [
          {
            name: "text.left_title",
            type: "markdown",
            value: "## Left",
          },
          {
            name: "chart.left_barchart",
            type: "bar",
            data: [
              ["Type", "Count"],
              ["User", 12],
              ["Policy", 93],
              ["Role", 48],
            ],
            title: "AWS IAM Entities",
          },
        ],
      },
      {
        name: "container.right",
        width: 6,
        children: [
          {
            name: "text.right_title",
            type: "markdown",
            value: "## Right",
          },
          {
            name: "chart.right_barchart",
            type: "line",
            data: [
              ["Type", "Count"],
              ["User", 12],
              ["Policy", 93],
              ["Role", 48],
            ],
            title: "AWS IAM Entities",
          },
        ],
      },
    ],
  },
};
