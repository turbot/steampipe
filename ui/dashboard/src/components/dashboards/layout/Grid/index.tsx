import { classNames } from "../../../../utils/styles";
import { Dispatch, MouseEventHandler, ReactNode, SetStateAction } from "react";
import { getResponsivePanelWidthClass } from "../../../../utils/layout";

interface GridEvents {
  [name: string]: MouseEventHandler<HTMLDivElement>;
}

type GridProps = {
  children: ReactNode;
  className?: string;
  name: string;
  width?: number;
  events?: GridEvents;
  setRef?: Dispatch<SetStateAction<null>>;
};

const Grid = ({
  children,
  className,
  name,
  width,
  events,
  setRef,
}: GridProps) => (
  <div
    // @ts-ignore
    ref={setRef}
    id={name}
    className={classNames(
      className,
      "grid grid-cols-12 col-span-12 gap-x-4 gap-y-4 md:gap-y-6 auto-rows-min",
      getResponsivePanelWidthClass(width)
    )}
    {...events}
  >
    {children}
  </div>
);

export default Grid;
