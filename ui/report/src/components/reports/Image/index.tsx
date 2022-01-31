import Img from "react-cool-img";
import { BasePrimitiveProps } from "../common";

export type ImageProps = BasePrimitiveProps & {
  properties: {
    src: string;
    alt: string;
  };
};

const Image = (props: ImageProps) => {
  return <Img src={props.properties.src} alt={props.properties.alt} />;
};

export default Image;
