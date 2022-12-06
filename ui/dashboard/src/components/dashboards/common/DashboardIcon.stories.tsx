import DashboardIcon from "./DashboardIcon";
import { ComponentMeta, ComponentStory } from "@storybook/react";

export default {
  title: "Dashboard Icon",
  component: DashboardIcon,
} as ComponentMeta<typeof DashboardIcon>;

const Template: ComponentStory<typeof DashboardIcon> = (args) => (
  <div className="h-full overflow-y-auto p-4">
    <DashboardIcon {...args} />
  </div>
);

export const heroIconDefaultOutline = Template.bind({});
heroIconDefaultOutline.args = {
  icon: "arrow-up-circle",
};

export const heroIconDefaultOutlineWithColor = Template.bind({});
heroIconDefaultOutlineWithColor.args = {
  icon: "arrow-up-circle",
  style: { color: "red" },
};

export const heroIconOutlineFullyNamespaced = Template.bind({});
heroIconOutlineFullyNamespaced.args = {
  icon: "heroicons-outline:arrow-up-circle",
};

export const heroIconOutlineFullyNamespacedWithColor = Template.bind({});
heroIconOutlineFullyNamespacedWithColor.args = {
  icon: "heroicons-outline:arrow-up-circle",
  style: { color: "red" },
};

export const heroIconSolid = Template.bind({});
heroIconSolid.args = {
  icon: "heroicons-solid:arrow-up-circle",
};

export const heroIconSolidWithColor = Template.bind({});
heroIconSolidWithColor.args = {
  icon: "heroicons-solid:arrow-up-circle",
  style: { color: "red" },
};

export const materialSymbolDefaultOutline = Template.bind({});
materialSymbolDefaultOutline.args = {
  icon: "cloud",
};

export const materialSymbolDefaultOutlineWithColor = Template.bind({});
materialSymbolDefaultOutlineWithColor.args = {
  icon: "cloud",
  style: { color: "red" },
};

export const materialSymbolOutlineFullyNamespaced = Template.bind({});
materialSymbolOutlineFullyNamespaced.args = {
  icon: "materialsymbols-outline:cloud",
};

export const materialSymbolOutlineFullyNamespacedWithColor = Template.bind({});
materialSymbolOutlineFullyNamespacedWithColor.args = {
  icon: "materialsymbols-outline:cloud",
  style: { color: "red" },
};

export const materialSymbolSolid = Template.bind({});
materialSymbolSolid.args = {
  icon: "materialsymbols-solid:cloud",
};

export const materialSymbolSolidWithColor = Template.bind({});
materialSymbolSolidWithColor.args = {
  icon: "materialsymbols-solid:cloud",
  style: { color: "red" },
};

export const text1Letter = Template.bind({});
text1Letter.args = {
  icon: "text:A",
};

export const text2Letter = Template.bind({});
text2Letter.args = {
  icon: "text:AB",
};

export const text3Letter = Template.bind({});
text3Letter.args = {
  icon: "text:ABC",
};

export const text4Letter = Template.bind({});
text4Letter.args = {
  icon: "text:ABCD",
};

export const text5Letter = Template.bind({});
text5Letter.args = {
  icon: "text:ABCDE",
};

export const text6Letter = Template.bind({});
text6Letter.args = {
  icon: "text:ABCDEF",
};

export const textSpaces = Template.bind({});
textSpaces.args = {
  icon: "text:Foo Bar",
};

export const SVGDataURL = Template.bind({});
SVGDataURL.args = {
  icon: "data:image/svg+xml;base64,PHN2ZyBjbGFzcz0idy02IGgtNiIgaGVpZ2h0PSI0OCIgd2lkdGg9IjQ4IiB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciPjxwYXRoIGQ9Ik00NC4wODMgMzIuNTgxYTcyLjg1MyA3Mi44NTMgMCAwMC0yLjA4Mi4xOTFsLS4xNjguMDE4Yy0xLjc2NS4xODItMy43NjMuMzg3LTQuNzE0LjI1Ny0uNTk0LS4wODUtMS42ODgtLjU2LTIuODQ4LTEuMDYzLTEuOTY2LS44NTItNC40MTQtMS45MTQtNi42MjctMi4wODQtMy4wNDYtLjIzNC04LjMxLjM0LTExLjcwMyAxLjkzMS0xLjk2Mi0uMDg4LTMuNzgyLS42MTYtNS4xOTYtMS40ODhhNDYuNDQgNDYuNDQgMCAwMS0uMDEyLS42NDljMC00LjEzNiAxLjkzOC04LjAxNSA1LjMyLTEwLjY1YTE1LjE3NiAxNS4xNzYgMCAwMTcuNjYtMi4wNTJjNi4xMiAwIDExLjU1MyAzLjYxNyAxMy42NzIgOC45NDJoLS44MTNjLS44MTQgMC0xLjQ5Ni41NzYtMS42NjMgMS4zNGwtLjQxMi0uMTU2Yy0yLjAzLS43ODMtNC44MTMtMS44NTQtOC4wMzEtMS42NjMtNC43MDUuMjc2LTcuMDIuNTYtOC44NDMgMS4wOGwuNTQ5IDEuOTI0YzEuNjY4LS40NzcgMy44NjgtLjc0IDguNDEtMS4wMDggMi43ODUtLjE1OCA1LjMzOC44MTggNy4xOTUgMS41MzMuMzk1LjE1My43NTIuMjg2IDEuMDkuNDA3di4yMDVjMCAuOTM5Ljc2NiAxLjcwNCAxLjcwNSAxLjcwNGgzLjY0N2MuOTM3IDAgMS43LS43NjIgMS43MDQtMS42OThsLjM1NC0uMDMyYTE3NS41NyAxNzUuNTcgMCAwMDEuNTk1LS4xNTFsLjIxLS4wMjF2My4xODN6TTkuNzk5IDM0LjI0OWMtMi42NS4zNjMtNC43NDIuMDM5LTUuNjA4LS40OTYuNTI1LS4zNzEgMS43MS0uODkgMi41NDQtMS4yNTRhNDMuNDgyIDQzLjQ4MiAwIDAwMi4xNi0xLjAwNGMxLjE1NC44NzkgMi41ODMgMS41MzggNC4xNjUgMS45MzEtLjkxNS4zNjgtMi4wNDguNjU4LTMuMjYxLjgyM3pNMzYuODY4IDI5LjNoMy4wNTV2LTEuMzY2aC0zLjA1NVYyOS4zek0yNi41ODYgMTNjOC4zNjggMCAxNS4zOTYgNi4yOTUgMTYuNDMgMTQuNDkyLS4yODcuMDI4LS41OTEuMDU3LS45MTIuMDg2bC0uMTg1LjAxNmExLjcwNCAxLjcwNCAwIDAwLTEuNy0xLjY2aC0uNzExYy0yLjIwMS02LjQ3Ny04LjU3Ni0xMC45NDItMTUuNzk1LTEwLjk0Mi0yLjY3NCAwLTUuMjkuNjEyLTcuNjQxIDEuNzc1LjAxNy0uMDE0LjAzLS4wMjkuMDQ2LS4wNDJBMTYuNTkgMTYuNTkgMCAwMTI2LjU4NiAxM3ptMTguOTQgMTQuNzk0YTEuNjI1IDEuNjI1IDAgMDAtLjQ5Ny0uMjk4QzQzLjk4MyAxOC4xODUgMzYuMDQ5IDExIDI2LjU4NiAxMWExOC41OTUgMTguNTk1IDAgMDAtMTEuNzMyIDQuMTc1IDE4LjUwNCAxOC41MDQgMCAwMC02Ljg1IDE0LjQwNmMwIC4wNDEuMDA0LjA4Mi4wMDQuMTIzLS42NjcuMzM1LTEuMzcyLjY1NS0yLjA3My45NjItMi4xMzYuOTM0LTMuNjggMS42MDgtMy45MSAyLjgyNi0uMDU3LjMwOS0uMDY3LjkxNC41MiAxLjUwMiAxLjAzOCAxLjAzNyAzLjAyOCAxLjQwOSA1LjA1IDEuNDA5Ljg0MiAwIDEuNjg4LS4wNjQgMi40NzQtLjE3MiAxLjMzOS0uMTgyIDMuODM5LS42NzMgNS41MzYtMS45MTUgMi42NzgtMS45NDEgOC43MzMtMi42NjcgMTEuODg2LTIuNDIyIDEuODc5LjE0NSA0LjE1NSAxLjEzMiA1Ljk4NSAxLjkyNSAxLjM5Ni42MDcgMi41IDEuMDg1IDMuMzY2IDEuMjA4IDEuMTkuMTYzIDMuMTM2LS4wMzcgNS4xOTUtLjI0OGwuMTczLS4wMTdhNjIuNjI3IDYyLjYyNyAwIDAxMS41OTItLjE1MWMuMjE1LS4wMjEuNDI4LS4wMzYuNzEyLS4wNTYuODgtLjA2MiAxLjU2OS0uOCAxLjU2OS0xLjY4di0zLjgyOWMwLS40NzgtLjIwMi0uOTM0LS41NTYtMS4yNTJ6IiBmaWxsPSIjQkYwODE2IiBmaWxsLXJ1bGU9ImV2ZW5vZGQiPjwvcGF0aD48L3N2Zz4=",
};
