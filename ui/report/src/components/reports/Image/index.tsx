import Img from "react-cool-img";
import Link from "../Link";
import {
  BasePrimitiveProps,
  ExecutablePrimitiveProps,
  LeafNodeData,
} from "../common";
import { get } from "lodash";
import { getColumnIndex } from "../../../utils/data";
import { useEffect, useState } from "react";

export type ImageProps = BasePrimitiveProps &
  ExecutablePrimitiveProps & {
    properties: {
      src: string;
      alt: string;
    };
  };

type ImageDataFormat = "simple" | "formal";

interface ImageState {
  loading: boolean;
  src: string | null;
  alt: string | null;
  link_url: string | null;
}

const getDataFormat = (data: LeafNodeData): ImageDataFormat => {
  if (data.columns.length > 1) {
    return "formal";
  }
  return "simple";
};

const useImageState = ({ data, properties }: ImageProps) => {
  const [calculatedProperties, setCalculatedProperties] = useState<ImageState>({
    loading: true,
    src: null,
    alt: null,
    link_url: null,
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
        loading: false,
        src: null,
        alt: null,
        link_url: null,
      });
      return;
    }

    const dataFormat = getDataFormat(data);

    if (dataFormat === "simple") {
      const firstCol = data.columns[0];
      const row = data.rows[0];
      setCalculatedProperties({
        loading: false,
        src: row[0],
        alt: firstCol.name,
        link_url: null,
      });
    } else {
      const altColIndex = getColumnIndex(data.columns, "alt");
      const formalAlt =
        altColIndex >= 0 ? get(data, `rows[0][${altColIndex}]`) : null;
      const srcColIndex = getColumnIndex(data.columns, "src");
      const formalSrc =
        srcColIndex >= 0 ? get(data, `rows[0][${srcColIndex}]`) : null;
      const linkUrlColIndex = getColumnIndex(data.columns, "link_url");
      const formalLinkUrl =
        linkUrlColIndex >= 0 ? get(data, `rows[0][${linkUrlColIndex}]`) : null;
      setCalculatedProperties({
        loading: false,
        src: formalSrc,
        alt: formalAlt,
        link_url: formalLinkUrl,
      });
    }
  }, [data, properties]);

  return calculatedProperties;
};

const Image = (props: ImageProps) => {
  const state = useImageState(props);
  const image = <Img src={state.src} alt={state.alt} />;
  if (state.link_url) {
    return <Link link_url={state.link_url}>{image}</Link>;
  }
  return image;
};

export default Image;
