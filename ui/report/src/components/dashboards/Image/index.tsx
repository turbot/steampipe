import Img from "react-cool-img";
import { BasePrimitiveProps, ExecutablePrimitiveProps } from "../common";

export type ImageProps = BasePrimitiveProps &
  ExecutablePrimitiveProps & {
    properties: {
      src: string;
      alt: string;
    };
  };

const Image = (props: ImageProps) => {
  return <Img src={props.properties.src} alt={props.properties.alt} />;
};

export default Image;
