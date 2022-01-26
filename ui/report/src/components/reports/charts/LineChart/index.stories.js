import LineChart from "./index";
import { PanelStoryDecorator } from "../../../../utils/storybook";

const story = {
  title: "Charts/Line",
  component: LineChart.component,
};

export default story;

const Template = (args) => (
  <PanelStoryDecorator
    definition={args}
    nodeType="chart"
    additionalProperties={{ type: "line" }}
  />
);

export const Loading = Template.bind({});
Loading.args = {
  loading: true,
};

export const Error = Template.bind({});
Error.args = {
  loading: false,
  error: "Something went wrong!",
};

export const Simple = Template.bind({});
Simple.args = {
  data: {
    columns: [
      { name: "Date", data_type_name: "TEXT" },
      { name: "Cost ($)", data_type_name: "INT8" },
    ],
    items: [
      { Date: "2020-02-01", "Cost ($)": 84.45 },
      { Date: "2020-02-02", "Cost ($)": 92.23 },
      { Date: "2020-02-03", "Cost ($)": 101.3 },
      { Date: "2020-02-04", "Cost ($)": 110.5 },
      { Date: "2020-02-05", "Cost ($)": 174.95 },
      { Date: "2020-02-06", "Cost ($)": 130.23 },
      { Date: "2020-02-07", "Cost ($)": 150.2 },
      { Date: "2020-02-08", "Cost ($)": 160 },
      { Date: "2020-02-09", "Cost ($)": 172.12 },
      { Date: "2020-02-10", "Cost ($)": 190 },
    ],
  },
};

export const LargeSeries = Template.bind({});
LargeSeries.args = {
  data: {
    columns: [
      { name: "Region", data_type_name: "TEXT" },
      { name: "Total", data_type_name: "INT8" },
    ],
    items: [
      { Region: "us-east-1", Total: 14 },
      { Region: "eu-central-1", Total: 6 },
      { Region: "ap-south-1", Total: 4 },
      { Region: "ap-southeast-1", Total: 3 },
      { Region: "ap-southeast-2", Total: 2 },
      { Region: "ca-central-1", Total: 2 },
      { Region: "eu-north-1", Total: 2 },
      { Region: "eu-west-1", Total: 1 },
      { Region: "eu-west-2", Total: 1 },
      { Region: "eu-west-3", Total: 1 },
      { Region: "sa-east-1", Total: 1 },
      { Region: "us-east-2", Total: 1 },
      { Region: "us-west-1", Total: 1 },
      { Region: "ap-northeast-1", Total: 1 },
      { Region: "us-west-2", Total: 1 },
      { Region: "ap-northeast-2", Total: 1 },
    ],
  },
};

export const MultiSeries = Template.bind({});
MultiSeries.storyName = "Multi-Series";
MultiSeries.args = {
  data: {
    columns: [
      { name: "Country", data_type_name: "TEXT" },
      { name: "Men", data_type_name: "INT8" },
      { name: "Women", data_type_name: "INT8" },
      { name: "Children", data_type_name: "INT8" },
    ],
    items: [
      { Country: "England", Men: 16000000, Women: 13000000, Children: 8000000 },
      { Country: "Scotland", Men: 8000000, Women: 7000000, Children: 3000000 },
      { Country: "Wales", Men: 5000000, Women: 3000000, Children: 2500000 },
      {
        Country: "Northern Ireland",
        Men: 3000000,
        Women: 2000000,
        Children: 1000000,
      },
    ],
  },
};

export const MultiSeriesOverrides = Template.bind({});
MultiSeriesOverrides.storyName = "Multi-Series with Series Overrides";
MultiSeriesOverrides.args = {
  data: {
    columns: [
      { name: "Country", data_type_name: "TEXT" },
      { name: "Men", data_type_name: "INT8" },
      { name: "Women", data_type_name: "INT8" },
      { name: "Children", data_type_name: "INT8" },
    ],
    items: [
      { Country: "England", Men: 16000000, Women: 13000000, Children: 8000000 },
      { Country: "Scotland", Men: 8000000, Women: 7000000, Children: 3000000 },
      { Country: "Wales", Men: 5000000, Women: 3000000, Children: 2500000 },
      {
        Country: "Northern Ireland",
        Men: 3000000,
        Women: 2000000,
        Children: 1000000,
      },
    ],
  },
  properties: {
    series: {
      Children: {
        title: "Kids",
        color: "green",
      },
    },
  },
};

export const SingleSeriesLegend = Template.bind({});
SingleSeriesLegend.storyName = "Single Series with Legend";
SingleSeriesLegend.args = {
  data: {
    columns: [
      { name: "Type", data_type_name: "TEXT" },
      { name: "Count", data_type_name: "INT8" },
    ],
    items: [
      { Type: "User", Count: 12 },
      { Type: "Policy", Count: 93 },
      { Type: "Role", Count: 48 },
    ],
  },
  properties: {
    legend: {
      display: "always",
    },
  },
};

export const SingleSeriesLegendPosition = Template.bind({});
SingleSeriesLegendPosition.storyName = "Single Series With Legend At Bottom";
SingleSeriesLegendPosition.args = {
  data: {
    columns: [
      { name: "Type", data_type_name: "TEXT" },
      { name: "Count", data_type_name: "INT8" },
    ],
    items: [
      { Type: "User", Count: 12 },
      { Type: "Policy", Count: 93 },
      { Type: "Role", Count: 48 },
    ],
  },
  properties: {
    legend: {
      display: "always",
      position: "bottom",
    },
  },
};

export const SingleSeriesXAxisTitle = Template.bind({});
SingleSeriesXAxisTitle.storyName = "Single Series with X Axis Title";
SingleSeriesXAxisTitle.args = {
  data: {
    columns: [
      { name: "Type", data_type_name: "TEXT" },
      { name: "Count", data_type_name: "INT8" },
    ],
    items: [
      { Type: "User", Count: 12 },
      { Type: "Policy", Count: 93 },
      { Type: "Role", Count: 48 },
    ],
  },
  properties: {
    axes: {
      x: {
        title: {
          display: "always",
          value: "I am a the X Axis title",
        },
      },
    },
  },
};

export const SingleSeriesXAxisNoLabels = Template.bind({});
SingleSeriesXAxisNoLabels.storyName = "Single Series with no X Axis Labels";
SingleSeriesXAxisNoLabels.args = {
  data: {
    columns: [
      { name: "Type", data_type_name: "TEXT" },
      { name: "Count", data_type_name: "INT8" },
    ],
    items: [
      { Type: "User", Count: 12 },
      { Type: "Policy", Count: 93 },
      { Type: "Role", Count: 48 },
    ],
  },
  properties: {
    axes: {
      x: {
        labels: {
          display: "none",
        },
      },
    },
  },
};

export const SingleSeriesYAxisNoLabels = Template.bind({});
SingleSeriesYAxisNoLabels.storyName = "Single Series with no Y Axis Labels";
SingleSeriesYAxisNoLabels.args = {
  data: {
    columns: [
      { name: "Type", data_type_name: "TEXT" },
      { name: "Count", data_type_name: "INT8" },
    ],
    items: [
      { Type: "User", Count: 12 },
      { Type: "Policy", Count: 93 },
      { Type: "Role", Count: 48 },
    ],
  },
  properties: {
    axes: {
      y: {
        labels: {
          display: "none",
        },
      },
    },
  },
};
