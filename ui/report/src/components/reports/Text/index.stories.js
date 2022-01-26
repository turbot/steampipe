import Text from "./index";
import { PanelStoryDecorator } from "../../../utils/storybook";

const story = {
  title: "Primitives/Text",
  component: Text,
};

export default story;

const Template = (args) => (
  <PanelStoryDecorator definition={args} nodeType="text" />
);

export const markdownWithTitle = Template.bind({});
markdownWithTitle.args = {
  title: "Injected title for the text",
  properties: { value: "This is the original text" },
};

export const markdownCode = Template.bind({});
markdownCode.args = {
  properties: {
    value: `\`\`\`
10 PRINT "Steampipe is great"
20 GOTO 10 
\`\`\``,
  },
};

export const markdownEmphasis = Template.bind({});
markdownEmphasis.args = {
  properties: { value: `_Lean_ machine` },
};

export const markdownLink = Template.bind({});
markdownLink.args = {
  properties: { value: "[Click to go somewhere](https://foo.bar)" },
};

export const markdownHeading1 = Template.bind({});
markdownHeading1.args = {
  properties: { value: "# Welcome to Steampipe Reports" },
};

export const markdownHeading2 = Template.bind({});
markdownHeading2.args = {
  properties: { value: "## Welcome to Steampipe Reports" },
};

export const markdownHeading3 = Template.bind({});
markdownHeading3.args = {
  properties: { value: "### Welcome to Steampipe Reports" },
};

export const markdownHeading4 = Template.bind({});
markdownHeading4.args = {
  properties: { value: "#### Welcome to Steampipe Reports" },
};

export const markdownHeading5 = Template.bind({});
markdownHeading5.args = {
  properties: { value: "##### Welcome to Steampipe Reports" },
};

export const markdownHeading6 = Template.bind({});
markdownHeading6.args = {
  properties: { value: "###### Welcome to Steampipe Reports" },
};

export const markdownHorizontalRule = Template.bind({});
markdownHorizontalRule.args = {
  properties: {
    value: `Above the fold
***
Below the fold`,
  },
};

export const markdownImage = Template.bind({});
markdownImage.args = {
  properties: {
    value: `![Steampipe](https://steampipe.io/images/steampipe_logo_wordmark_color.svg)`,
  },
};

export const markdownOrderedList = Template.bind({});
markdownOrderedList.args = {
  properties: {
    value: `1. foo
2. bar`,
  },
};

export const markdownUnorderedList = Template.bind({});
markdownUnorderedList.args = {
  properties: {
    value: `- foo
- bar`,
  },
};

export const markdownParagraph = Template.bind({});
markdownParagraph.args = {
  properties: { value: "I am a paragraph of text" },
};

export const markdownParagraphWithInlineCode = Template.bind({});
markdownParagraphWithInlineCode.args = {
  properties: { value: "I am a paragraph of text containing `inline code`" },
};

export const markdownStrikethrough = Template.bind({});
markdownStrikethrough.args = {
  properties: { value: `Just ~~cross out~~ the bits you don't want` },
};

export const markdownStrong = Template.bind({});
markdownStrong.args = {
  properties: { value: `**Bold** is best` },
};

export const markdownTable = Template.bind({});
markdownTable.args = {
  properties: {
    value: `| Name | Company | Position |
| ----- | ----- | ----- |
| David Brent | Wernham Hogg | General Manager |
| Michael Scott | Dunder Mifflin | Regional Manager |`,
  },
};

export const markdownKitchenSink = Template.bind({});
markdownKitchenSink.args = {
  properties: {
    value: `# Start with a heading

Then some intro text

## Followed by a sub-heading

Then we'll [link](https://some.where), **bold** some text, _emphasise_ some text, or perhaps even *__both__*.

Todo:

- Write more components
- Drink more coffee
- Try \`inline\` code

***

Finally, here's a poem:

\`\`\`
Shall I compare thee to a summer’s day?
Thou art more lovely and more temperate.
Rough winds do shake the darling buds of May,
And summer’s lease hath all too short a date.
Sometime too hot the eye of heaven shines,
And often is his gold complexion dimmed;
And every fair from fair sometime declines,
By chance, or nature’s changing course, untrimmed;
But thy eternal summer shall not fade,
Nor lose possession of that fair thou ow’st,
Nor shall death brag thou wand’rest in his shade,
When in eternal lines to Time thou grow’st.
So long as men can breathe, or eyes can see,
So long lives this, and this gives life to thee.
\`\`\``,
  },
};

export const rawWithMarkdown = Template.bind({});
rawWithMarkdown.args = {
  properties: {
    value: `## Title goes here

**Bold** is best`,
    type: "raw",
  },
};
