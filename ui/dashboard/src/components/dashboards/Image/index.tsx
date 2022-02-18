import Img from "react-cool-img";
import Table from "../Table";
import { BasePrimitiveProps, ExecutablePrimitiveProps } from "../common";
import { get } from "lodash";

type ImageType = "image" | "table" | null;

export type ImageProps = BasePrimitiveProps &
  ExecutablePrimitiveProps & {
    properties: {
      type?: ImageType;
      src: string;
      alt: string;
    };
  };

const Image = (props: ImageProps) => {
  if (get(props, "properties.type") === "table") {
    // @ts-ignore
    return <Table {...props} />;
  }
  return <Img src={props.properties.src} alt={props.properties.alt} />;
};

export default Image;
