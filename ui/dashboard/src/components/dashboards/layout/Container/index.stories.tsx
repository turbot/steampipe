import Container from "./index";
import { PanelStoryDecorator } from "../../../../utils/storybook";

const story = {
  title: "Layout/Container",
  component: Container,
};

export default story;

const Template = (args) => (
  <PanelStoryDecorator
    definition={args.definition}
    panels={args.panels}
    panelType="container"
  />
);

export const Empty = Template.bind({});
// @ts-ignore
Empty.args = {
  definition: {},
};

export const Basic = Template.bind({});
const textPanel = {
  name: "text.markdown",
  panel_type: "text",
  properties: { value: "## Basic Dashboard" },
  status: "complete",
};
// @ts-ignore
Basic.args = {
  definition: {
    children: [textPanel],
    panel_type: "container",
    title: "Basic Container",
  },
  panels: {
    [textPanel.name]: textPanel,
  },
};

export const TwoColumn = Template.bind({});
const leftContainer = {
  name: "container.left",
  panel_type: "container",
  width: 6,
};
const rightContainer = {
  name: "container.right",
  panel_type: "container",
  width: 6,
};
const leftTopTextPanel = {
  name: "left.top.text.markdown",
  panel_type: "text",
  properties: { value: "## Column 1 Top" },
  status: "complete",
};
const leftBottomTextPanel = {
  name: "left.bottom.text.markdown",
  panel_type: "text",
  properties: { value: "## Column 1 Bottom" },
  status: "complete",
};
const rightTopTextPanel = {
  name: "right.top.text.markdown",
  panel_type: "text",
  properties: { value: "## Column 2 Top" },
  status: "complete",
};
const rightBottomTextPanel = {
  name: "right.bottom.text.markdown",
  panel_type: "text",
  properties: { value: "## Column 2 Bottom" },
  status: "complete",
};
// @ts-ignore
TwoColumn.args = {
  definition: {
    children: [
      {
        ...leftContainer,
        children: [leftTopTextPanel, leftBottomTextPanel],
      },
      {
        ...rightContainer,
        children: [rightTopTextPanel, rightBottomTextPanel],
      },
    ],
  },
  panels: {
    [leftContainer.name]: leftContainer,
    [rightContainer.name]: rightContainer,
    [leftTopTextPanel.name]: leftTopTextPanel,
    [leftBottomTextPanel.name]: leftBottomTextPanel,
    [rightTopTextPanel.name]: rightTopTextPanel,
    [rightBottomTextPanel.name]: rightBottomTextPanel,
  },
};
