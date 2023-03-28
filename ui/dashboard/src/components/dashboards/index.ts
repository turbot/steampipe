import { ReactNode } from "react";

const componentsMap = {};

const getComponent = (key: string) => componentsMap[key];

const registerComponent = (
  key: string,
  component: (props: any) => ReactNode
) => {
  componentsMap[key] = component;
};

export { getComponent, registerComponent };
