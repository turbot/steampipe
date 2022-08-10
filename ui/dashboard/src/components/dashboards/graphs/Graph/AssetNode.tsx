import { classNames } from "../../../../utils/styles";
import { createPortal } from "react-dom";
import { Handle } from "react-flow-renderer";
import { forwardRef, memo, useState } from "react";
import { ThemeProvider, ThemeWrapper } from "../../../../hooks/useTheme";
import { usePopper } from "react-popper";

interface TooltipProps {
  children: JSX.Element | JSX.Element[];
  styles: {};
  title: string;
}

const Tooltip = forwardRef(
  ({ children, styles, title, ...rest }: TooltipProps, ref) => {
    return (
      <ThemeProvider>
        <ThemeWrapper>
          <div
            // @ts-ignore
            ref={ref}
            style={styles}
            {...rest}
            className="z-50 bg-dashboard p-3 border border-divide rounded-md text-sm flex flex-col space-y-2 bg-dashboard-panel"
          >
            <Title title={title} />
            {children}
          </div>
        </ThemeWrapper>
      </ThemeProvider>
    );
  }
);

const Title = ({ title }) => {
  return <strong className="block break-all">{title}</strong>;
};

const PropertyItem = ({ name, value }) => {
  return (
    <div>
      <span className="block text-sm text-table-head truncate">{name}</span>
      {value === null && (
        <span className="text-foreground-lightest">
          <>null</>
        </span>
      )}
      {value !== null && (
        <span className={classNames("block", "break-words")}>{value}</span>
      )}
    </div>
    // <div className="space-x-2">
    //   <span>{name}</span>
    //   <span>=</span>
    //   <span>{value}</span>
    // </div>
  );
};

const Properties = ({ properties = {} }) => {
  return (
    <div className="space-y-2">
      {Object.entries(properties || {}).map(([key, value]) => (
        <PropertyItem key={key} name={key} value={value} />
      ))}
    </div>
  );
};

const AssetNode = ({ data }) => {
  const [showProperties, setShowProperties] = useState(false);
  const [referenceElement, setReferenceElement] = useState(null);
  const [popperElement, setPopperElement] = useState(null);
  const [arrowElement, setArrowElement] = useState(null);
  const { styles, attributes } = usePopper(referenceElement, popperElement, {
    modifiers: [{ name: "arrow", options: { element: arrowElement } }],
  });
  const icon = data.icon ? data.icon : null;

  // Notes:
  // * The Handle elements seem to be required to allow the connectors to work.
  return (
    <>
      {/*@ts-ignore*/}
      <Handle type="target" />
      {/*@ts-ignore*/}
      <Handle type="source" />
      <div
        className="flex flex-col items-center cursor-auto"
        // @ts-ignore
        ref={setReferenceElement}
      >
        <div
          className="py-2 px-2 rounded-full w-[35px] h-[35px] text-sm leading-[35px] my-0 mx-auto border border-divide cursor-grab"
          onMouseEnter={() => setShowProperties(true)}
          onMouseLeave={() => setShowProperties(false)}
        >
          {icon && <img className="max-w-full" src={icon} />}
        </div>
        <div className="text-[7px] mt-1 bg-dashboard-panel text-foreground min-w-[35px]">
          {data.label}
        </div>
      </div>
      {showProperties &&
        createPortal(
          // <div
          //   // @ts-ignore
          //   ref={setPopperElement}
          //   style={styles.popper}
          //   {...attributes.popper}
          // >
          <Tooltip
            ref={setPopperElement}
            styles={styles.popper}
            title={data.label}
            {...attributes.popper}
          >
            <Properties properties={data.properties} />
          </Tooltip>,
          // @ts-ignore
          document.querySelector("body")
        )}
    </>
  );
};

export default memo(AssetNode);
