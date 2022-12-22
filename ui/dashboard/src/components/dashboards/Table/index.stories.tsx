import Table from "./index";
import { PanelStoryDecorator } from "../../../utils/storybook";
import { ComponentMeta, ComponentStory } from "@storybook/react";

export default {
  title: "Primitives/Table",
  component: Table,
} as ComponentMeta<typeof Table>;

const Template: ComponentStory<typeof Table> = (args) => (
  <PanelStoryDecorator definition={args} panelType="table" />
);

export const Loading = Template.bind({});
Loading.args = {
  // @ts-ignore
  data: null,
};

export const NoData = Template.bind({});
NoData.args = {
  // @ts-ignore
  data: {
    columns: [
      { name: "Region", data_type: "TEXT" },
      { name: "Resource Count", data_type: "INT8" },
    ],
  },
};

export const Basic = Template.bind({});
Basic.args = {
  data: {
    columns: [
      { name: "Region", data_type: "TEXT" },
      { name: "Resource Count", data_type: "INT8" },
      { name: "Encrypted", data_type: "BOOL" },
      { name: "Info", data_type: "JSONB" },
    ],
    rows: [
      {
        Region: "us-east-1",
        "Resource Count": 246,
        Encrypted: true,
        Info: null,
      },
      {
        Region: "us-east-2",
        "Resource Count": 146,
        Encrypted: false,
        Info: { foo: "bar" },
      },
      {
        Region: "us-west-1",
        "Resource Count": 57,
        Encrypted: true,
        Info: { bar: "foo" },
      },
      {
        Region: "eu-west-1",
        "Resource Count": 290,
        Encrypted: false,
        Info: { foobar: "barfoo" },
      },
      {
        Region: "eu-west-2",
        "Resource Count": 198,
        Encrypted: true,
        Info: null,
      },
    ],
  },
};

export const Nulls = Template.bind({});
Nulls.args = {
  data: {
    columns: [
      { name: "Color", data_type: "TEXT" },
      { name: "Value", data_type: "INT8" },
    ],
    rows: [
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

export const HideColumn = Template.bind({});
HideColumn.args = {
  data: {
    columns: [
      { name: "Region", data_type: "TEXT" },
      { name: "Resource Count", data_type: "INT8" },
      { name: "Encrypted", data_type: "BOOL" },
      { name: "Info", data_type: "JSONB" },
    ],
    rows: [
      {
        Region: "us-east-1",
        "Resource Count": 246,
        Encrypted: true,
        Info: null,
      },
      {
        Region: "us-east-2",
        "Resource Count": 146,
        Encrypted: false,
        Info: { foo: "bar" },
      },
      {
        Region: "us-west-1",
        "Resource Count": 57,
        Encrypted: true,
        Info: { bar: "foo" },
      },
      {
        Region: "eu-west-1",
        "Resource Count": 290,
        Encrypted: false,
        Info: { foobar: "barfoo" },
      },
      {
        Region: "eu-west-2",
        "Resource Count": 198,
        Encrypted: true,
        Info: null,
      },
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
      { name: "Region", data_type: "TEXT" },
      { name: "Resource Count", data_type: "INT8" },
      { name: "Encrypted", data_type: "BOOL" },
      { name: "Long", data_type: "TEXT" },
    ],
    rows: [
      {
        Region: "us-east-1",
        "Resource Count": 246,
        Encrypted: true,
        Long: "Some really, really, really, really, really, really, really, really, really, really, long text with the occasional realllllllllllllllllly, realllllllllllllllllly, realllllllllllllllllly, realllllllllllllllllly long word that we want to wrap in a better way.",
      },
      {
        Region: "us-east-2",
        "Resource Count": 146,
        Encrypted: false,
        Long: "Some really, really, really, really, really, really, really, really, really, really, long text with the occasional realllllllllllllllllly, realllllllllllllllllly, realllllllllllllllllly, realllllllllllllllllly long word that we want to wrap in a better way.",
      },
      {
        Region: "us-west-1",
        "Resource Count": 57,
        Encrypted: true,
        Long: "Some really, really, really, really, really, really, really, really, really, really, long text with the occasional realllllllllllllllllly, realllllllllllllllllly, realllllllllllllllllly, realllllllllllllllllly long word that we want to wrap in a better way.",
      },
      {
        Region: "eu-west-1",
        "Resource Count": 290,
        Encrypted: false,
        Long: "Some really, really, really, really, really, really, really, really, really, really, long text with the occasional realllllllllllllllllly, realllllllllllllllllly, realllllllllllllllllly, realllllllllllllllllly long word that we want to wrap in a better way.",
      },
      {
        Region: "eu-west-2",
        "Resource Count": 198,
        Encrypted: true,
        Long: "Some really, really, really, really, really, really, really, really, really, really, long text with the occasional realllllllllllllllllly, realllllllllllllllllly, realllllllllllllllllly, realllllllllllllllllly long word that we want to wrap in a better way.",
      },
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

export const LineView = Template.bind({});
LineView.args = {
  display_type: "line",
  data: {
    columns: [
      { name: "Region", data_type: "TEXT" },
      { name: "Resource Count", data_type: "INT8" },
      { name: "Encrypted", data_type: "BOOL" },
      { name: "Info", data_type: "JSONB" },
    ],
    rows: [
      {
        Region: "us-east-1",
        "Resource Count": 246,
        Encrypted: true,
        Info: null,
      },
    ],
  },
};

export const LineViewMultipleRows = Template.bind({});
LineViewMultipleRows.args = {
  display_type: "line",
  data: {
    columns: [
      { name: "Region", data_type: "TEXT" },
      { name: "Resource Count", data_type: "INT8" },
      { name: "Encrypted", data_type: "BOOL" },
      { name: "Info", data_type: "JSONB" },
    ],
    rows: [
      {
        Region: "us-east-1",
        "Resource Count": 246,
        Encrypted: true,
        Info: null,
      },
      {
        Region: "us-east-2",
        "Resource Count": 146,
        Encrypted: false,
        Info: { foo: "bar" },
      },
    ],
  },
};

export const LineViewHideColumn = Template.bind({});
LineViewHideColumn.args = {
  display_type: "line",
  data: {
    columns: [
      { name: "Region", data_type: "TEXT" },
      { name: "Resource Count", data_type: "INT8" },
      { name: "Encrypted", data_type: "BOOL" },
      { name: "Info", data_type: "JSONB" },
    ],
    rows: [
      {
        Region: "us-east-1",
        "Resource Count": 246,
        Encrypted: true,
        Info: null,
      },
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

export const LineViewWrapColumn = Template.bind({});
LineViewWrapColumn.args = {
  display_type: "line",
  data: {
    columns: [
      { name: "Region", data_type: "TEXT" },
      { name: "Resource Count", data_type: "INT8" },
      { name: "Encrypted", data_type: "BOOL" },
      { name: "Long", data_type: "TEXT" },
    ],
    rows: [
      {
        Region: "us-east-1",
        "Resource Count": 246,
        Encrypted: true,
        Long: "Some really, really, really, really, really, really, really, really, really, really, long text with the occasional realllllllllllllllllly, realllllllllllllllllly, realllllllllllllllllly, realllllllllllllllllly long word that we want to wrap in a better way.",
      },
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
