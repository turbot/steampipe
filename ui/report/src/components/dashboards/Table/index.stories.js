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
      { name: "Encrypted", data_type_name: "BOOL" },
      { name: "Info", data_type_name: "JSONB" },
    ],
    rows: [
      ["us-east-1", 246, true, null],
      ["us-east-2", 146, false, { foo: "bar" }],
      ["us-west-1", 57, true, { bar: "foo" }],
      ["eu-west-1", 290, false, { foobar: "barfoo" }],
      ["eu-west-2", 198, true, null],
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

export const HideColumn = Template.bind({});
HideColumn.args = {
  data: {
    columns: [
      { name: "Region", data_type_name: "TEXT" },
      { name: "Resource Count", data_type_name: "INT8" },
      { name: "Encrypted", data_type_name: "BOOL" },
      { name: "Info", data_type_name: "JSONB" },
    ],
    rows: [
      ["us-east-1", 246, true, null],
      ["us-east-2", 146, false, { foo: "bar" }],
      ["us-west-1", 57, true, { bar: "foo" }],
      ["eu-west-1", 290, false, { foobar: "barfoo" }],
      ["eu-west-2", 198, true, null],
    ],
  },
  properties: {
    columns: {
      Info: {
        display: "none",
      },
    },
  },
};

export const WrapColumn = Template.bind({});
WrapColumn.args = {
  data: {
    columns: [
      { name: "Region", data_type_name: "TEXT" },
      { name: "Resource Count", data_type_name: "INT8" },
      { name: "Encrypted", data_type_name: "BOOL" },
      { name: "Long", data_type_name: "TEXT" },
    ],
    rows: [
      [
        "us-east-1",
        246,
        true,
        "I am a really, really, really, really, really, really, really, really, really, really long value that I wish to wrap",
      ],
      [
        "us-east-2",
        146,
        false,
        "I am a really, really, really, really, really, really, really, really, really, really long value that I wish to wrap",
      ],
      [
        "us-west-1",
        57,
        true,
        "I am a really, really, really, really, really, really, really, really, really, really long value that I wish to wrap",
      ],
      [
        "eu-west-1",
        290,
        false,
        "I am a really, really, really, really, really, really, really, really, really, really long value that I wish to wrap",
      ],
      [
        "eu-west-2",
        198,
        true,
        "I am a really, really, really, really, really, really, really, really, really, really long value that I wish to wrap",
      ],
    ],
  },
  properties: {
    columns: {
      Long: {
        wrap: "all",
      },
    },
  },
};
