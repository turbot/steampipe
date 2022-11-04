import { classNames } from "../../../../utils/styles";
import { MouseEventHandler, ReactNode } from "react";
import { getResponsivePanelWidthClass } from "../../../../utils/layout";

interface GridEvents {
  [name: string]: MouseEventHandler<HTMLDivElement>;
}

type GridProps = {
  children: ReactNode;
  name: string;
  width?: number;
  events?: GridEvents;
};

const Grid = ({ children, name, width, events }: GridProps) => (
  <div
    id={name}
    className={classNames(
      "grid grid-cols-12 col-span-12 gap-x-4 gap-y-4 md:gap-y-6 auto-rows-min",
      getResponsivePanelWidthClass(width)
    )}
    {...events}
  >
    {children}
  </div>
);

export default Grid;
