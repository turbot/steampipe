import Container from "./index";
import { PanelStoryDecorator } from "../../../../utils/storybook";

const story = {
  title: "Layout/Container",
  component: Container,
};

export default story;

const Template = (args) => (
  <PanelStoryDecorator definition={args.definition} nodeType="container" />
);

export const Empty = Template.bind({});
// @ts-ignore
Empty.args = {
  definition: {},
};

export const Title = Template.bind({});
// @ts-ignore
Title.args = {
  definition: {
    title: "Container Title",
  },
};

export const Basic = Template.bind({});
// @ts-ignore
Basic.args = {
  definition: {
    children: [{ name: "text.markdown", value: "## Basic Report" }],
  },
};

export const TwoColumn = Template.bind({});
// @ts-ignore
TwoColumn.args = {
  definition: {
    children: [
      {
        name: "container.left",
        width: 6,
        children: [
          {
            name: "text.markdown",
            value: "## Column 1 Top",
          },
          {
            name: "text.markdown",
            value: "## Column 1 Bottom",
          },
        ],
      },
      {
        name: "container.right",
        width: 6,
        children: [
          {
            name: "text.markdown",
            value: "## Column 2 Top",
          },
          {
            name: "text.markdown",
            value: "## Column 2 Bottom",
          },
        ],
      },
    ],
  },
};
