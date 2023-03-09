import DonutChart from "./index";
import { PanelStoryDecorator } from "../../../../utils/storybook";

const story = {
  title: "Charts/Donut",
  component: DonutChart.component,
};

export default story;

const Template = (args) => (
  <PanelStoryDecorator
    definition={{ ...args, display_type: "donut" }}
    panelType="chart"
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

export const SingleSeries = Template.bind({});
SingleSeries.storyName = "Single Series";
SingleSeries.args = {
  data: {
    columns: [
      { name: "Type", data_type: "TEXT" },
      { name: "Count", data_type: "INT8" },
    ],
    rows: [
      { Type: "User", Count: 12 },
      { Type: "Policy", Count: 93 },
      { Type: "Role", Count: 48 },
    ],
  },
};

export const LargeSeries = Template.bind({});
LargeSeries.args = {
  data: {
    columns: [
      { name: "Region", data_type: "TEXT" },
      { name: "Total", data_type: "INT8" },
    ],
    rows: [
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
      { name: "Country", data_type: "TEXT" },
      { name: "Men", data_type: "INT8" },
      { name: "Women", data_type: "INT8" },
      { name: "Children", data_type: "INT8" },
    ],
    rows: [
      { Country: "England", Men: 16000000, Women: 13000000, Children: 8000000 },
      { Country: "Scotland", Men: 8000000, Women: 7000000, Children: 3000000 },
      {
        Country: "Wales",
        Men: 5000000,
        Women: 3000000,
        Children: 2500000,
      },
      {
        Country: "Northern Ireland",
        Men: 3000000,
        Women: 2000000,
        Children: 1000000,
      },
    ],
  },
};

export const PointColorOverrides = Template.bind({});
PointColorOverrides.args = {
  data: {
    columns: [
      { name: "versioning_status", data_type: "TEXT" },
      { name: "Total", data_type: "INT8" },
    ],
    rows: [
      { versioning_status: "Disabled", Total: 2 },
      { versioning_status: "Enabled", Total: 14 },
    ],
  },
  properties: {
    series: {
      Total: {
        points: {
          Disabled: { name: "Disabled", color: "red" },
          Enabled: { name: "Enabled", color: "green" },
        },
      },
    },
  },
};
