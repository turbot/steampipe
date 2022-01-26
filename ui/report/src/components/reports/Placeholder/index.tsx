import ReactPlaceholder from "react-placeholder";
import "react-placeholder/lib/reactPlaceholder.css";

type PlaceholderProps = {
  animate: boolean;
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
    >
      {children}
    </ReactPlaceholder>
  );
};

export default {
  type: "placeholder",
  component: Placeholder,
};
