import DonutChart from "./index";
import { PanelStoryDecorator } from "../../../../utils/storybook";

const story = {
  title: "Charts/Donut",
  component: DonutChart.component,
};

export default story;

const Template = (args) => (
  <PanelStoryDecorator
    definition={args}
    nodeType="chart"
    additionalProperties={{ type: "donut" }}
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
      { name: "Type", data_type_name: "TEXT" },
      { name: "Count", data_type_name: "INT8" },
    ],
    rows: [
      ["User", 12],
      ["Policy", 93],
      ["Role", 48],
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
    rows: [
      ["us-east-1", 14],
      ["eu-central-1", 6],
      ["ap-south-1", 4],
      ["ap-southeast-1", 3],
      ["ap-southeast-2", 2],
      ["ca-central-1", 2],
      ["eu-north-1", 2],
      ["eu-west-1", 1],
      ["eu-west-2", 1],
      ["eu-west-3", 1],
      ["sa-east-1", 1],
      ["us-east-2", 1],
      ["us-west-1", 1],
      ["ap-northeast-1", 1],
      ["us-west-2", 1],
      ["ap-northeast-2", 1],
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
    rows: [
      ["England", 16000000, 13000000, 8000000],
      ["Scotland", 8000000, 7000000, 3000000],
      ["Wales", 5000000, 3000000, 2500000],
      ["Northern Ireland", 3000000, 2000000, 1000000],
    ],
  },
};
