import Table from "./index";
import { PanelStoryDecorator } from "../../../utils/storybook";

const story = {
  title: "Primitives/Table",
  component: Table,
};

export default story;

const Template = (args) => (
  <PanelStoryDecorator definition={args} nodeType="table" />
);

export const Loading = Template.bind({});
Loading.args = {
  data: null,
};

export const NoData = Template.bind({});
NoData.args = {
  data: {
    columns: [
      { name: "Region", data_type_name: "TEXT" },
      { name: "Resource Count", data_type_name: "INT8" },
    ],
  },
};

export const Basic = Template.bind({});
Basic.args = {
  data: {
    columns: [
      { name: "Region", data_type_name: "TEXT" },
      { name: "Resource Count", data_type_name: "INT8" },
    ],
    items: [
      { Region: "us-east-1", "Resource Count": 246 },
      { Region: "us-east-2", "Resource Count": 146 },
      { Region: "us-west-1", "Resource Count": 57 },
      { Region: "eu-west-1", "Resource Count": 290 },
      { Region: "eu-west-2", "Resource Count": 198 },
    ],
  },
};

export const Nulls = Template.bind({});
Nulls.args = {
  data: {
    columns: [
      { name: "Color", data_type_name: "TEXT" },
      { name: "Value", data_type_name: "INT8" },
    ],
    items: [
      { Color: "red", Value: 10 },
      { Color: "orange", Value: null },
      { Color: "yellow", Value: 5 },
      { Color: "green", Value: null },
      { Color: "blue", Value: 2 },
      { Color: "indigo", Value: null },
      { Color: "violet", Value: 0 },
    ],
  },
};
