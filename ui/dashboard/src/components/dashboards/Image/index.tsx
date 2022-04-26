import get from "lodash/get";
import Img from "react-cool-img";
import Table from "../Table";
import {
  BasePrimitiveProps,
  ExecutablePrimitiveProps,
  LeafNodeData,
} from "../common";
import { getColumnIndex } from "../../../utils/data";
import { useEffect, useState } from "react";

type ImageType = "image" | "table" | null;

type ImageDataFormat = "simple" | "formal";

interface ImageState {
  src: string | null;
  alt: string | null;
}

export type ImageProps = BasePrimitiveProps &
  ExecutablePrimitiveProps & {
    properties: {
      type?: ImageType;
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
        src: row[0],
        alt: firstCol.name,
      });
    } else {
      const srcColIndex = getColumnIndex(data.columns, "src");
      const src =
        srcColIndex >= 0 ? get(data, `rows[0][${srcColIndex}]`) : null;
      const altColIndex = getColumnIndex(data.columns, "alt");
      const alt =
        altColIndex >= 0 ? get(data, `rows[0][${altColIndex}]`) : null;

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
  if (get(props, "properties.type") === "table") {
    // @ts-ignore
    return <Table {...props} />;
  }
  return <Image {...props} />;
};

export default ImageWrapper;
