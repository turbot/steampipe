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
    rows: [
      ["us-east-1", 246],
      ["us-east-2", 146],
      ["us-west-1", 57],
      ["eu-west-1", 290],
      ["eu-west-2", 198],
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
    rows: [
      ["red", 10],
      ["orange", null],
      ["yellow", 5],
      ["green", null],
      ["blue", 2],
      ["indigo", null],
      ["violet", 0],
    ],
  },
};
