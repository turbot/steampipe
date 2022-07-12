import get from "lodash/get";
import Img from "react-cool-img";
import {
  BasePrimitiveProps,
  ExecutablePrimitiveProps,
  LeafNodeData,
} from "../common";
import { getComponent, registerComponent } from "../index";
import { useEffect, useState } from "react";
const Table = getComponent("table");

type ImageType = "image" | "table" | null;

type ImageDataFormat = "simple" | "formal";

interface ImageState {
  src: string | null;
  alt: string | null;
}

export type ImageProps = BasePrimitiveProps &
  ExecutablePrimitiveProps & {
    display_type?: ImageType;
    properties: {
      src: string;
      alt: string;
    };
  };

const getDataFormat = (data: LeafNodeData): ImageDataFormat => {
  if (data.columns.length > 1) {
    return "formal";
  }
  return "simple";
};

const useImageState = ({ data, properties }: ImageProps) => {
  const [calculatedProperties, setCalculatedProperties] = useState<ImageState>({
    src: properties.src || null,
    alt: properties.alt || null,
  });

  useEffect(() => {
    if (!data) {
      return;
    }

    if (
      !data.columns ||
      !data.rows ||
      data.columns.length === 0 ||
      data.rows.length === 0
    ) {
      setCalculatedProperties({
        src: null,
        alt: null,
      });
      return;
    }

    const dataFormat = getDataFormat(data);

    if (dataFormat === "simple") {
      const firstCol = data.columns[0];
      const row = data.rows[0];
      setCalculatedProperties({
        src: row[firstCol.name],
        alt: firstCol.name,
      });
    } else {
      const src = get(data, `rows[0].src`, null);
      const alt = get(data, `rows[0].alt`, null);

      setCalculatedProperties({
        src,
        alt,
      });
    }
  }, [data, properties]);

  return calculatedProperties;
};

const Image = (props: ImageProps) => {
  const state = useImageState(props);
  return <Img src={state.src} alt={state.alt} />;
};

const ImageWrapper = (props: ImageProps) => {
  if (props.display_type === "table") {
    // @ts-ignore
    return <Table {...props} />;
  }
  return <Image {...props} />;
};

registerComponent("image", ImageWrapper);

export default ImageWrapper;
