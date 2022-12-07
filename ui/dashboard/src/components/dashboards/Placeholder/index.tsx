import ReactPlaceholder from "react-placeholder";
import "react-placeholder/lib/reactPlaceholder.css";

type PlaceholderProps = {
  animate?: boolean;
  children: null | JSX.Element | JSX.Element[];
  ready: boolean;
  type?: "rect" | "text" | "textRow" | "media" | "round" | undefined;
};

const Placeholder = ({
  animate = true,
  children,
  ready,
  type,
}: PlaceholderProps) => {
  return (
    // @ts-ignore
    <ReactPlaceholder
      ready={ready}
      rows={8}
      type={type}
      showLoadingAnimation={animate}
      style={{ padding: "1em" }}
    >
      {children}
    </ReactPlaceholder>
  );
};

const component = {
  type: "placeholder",
  component: Placeholder,
};

export default component;
