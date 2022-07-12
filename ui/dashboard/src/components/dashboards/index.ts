const componentsMap = {};

const getComponent = (key: string) => componentsMap[key];

const registerComponent = (
  key: string,
  component: (props: any) => JSX.Element | null
) => {
  componentsMap[key] = component;
};

export { getComponent, registerComponent };
