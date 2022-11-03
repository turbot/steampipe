import { classNames } from "../../../../utils/styles";
import { ReactNode } from "react";
import { getResponsivePanelWidthClass } from "../../../../utils/layout";

type GridProps = {
  children: ReactNode;
  name: string;
  width?: number;
};

const Grid = ({ children, name, width }: GridProps) => (
  <div
    id={name}
    className={classNames(
      "grid grid-cols-12 col-span-12 gap-x-4 gap-y-4 md:gap-y-6 auto-rows-min",
      getResponsivePanelWidthClass(width)
    )}
  >
    {children}
  </div>
);

export default Grid;
