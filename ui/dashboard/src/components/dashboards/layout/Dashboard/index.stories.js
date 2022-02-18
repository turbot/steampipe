import Dashboard from "./index";
import { DashboardContext } from "../../../../hooks/useDashboard";

const story = {
  title: "Layout/Dashboard",
  component: Dashboard,
};

export default story;

const Template = (args) => (
  <DashboardContext.Provider
    value={{ dispatch: () => {}, dashboard: args.dashboard }}
  >
    <Dashboard />
  </DashboardContext.Provider>
);

export const Basic = Template.bind({});
Basic.args = {
  dashboard: {
    name: "dashboard.basic",
    title: "Basic Dashboard",
    children: [
      {
        name: "text.header",
        type: "markdown",
        value: "## Basic Dashboard",
      },
    ],
  },
};

export const TwoColumn = Template.bind({});
TwoColumn.args = {
  dashboard: {
    name: "dashboard.two_column",
    title: "Two Column Dashboard",
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
  dashboard: {
    name: "dashboard.layout_container",
    title: "Layout Container Dashboard",
    children: [
      {
        name: "container.wrapper",
        children: [
          {
            name: "text.title",
            type: "markdown",
            value: "## IAM Dashboard",
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
  dashboard: {
    name: "dashboard.layout_container",
    title: "Layout Container Dashboard",
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
