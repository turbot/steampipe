import { cloneElement, useState } from "react";
import { createPortal } from "react-dom";
import { ThemeProvider, ThemeWrapper } from "../../../../hooks/useTheme";
import { usePopper } from "react-popper";

interface TooltipProps {
  children: JSX.Element;
  overlay?: JSX.Element | JSX.Element[];
  show?: boolean;
  title: string;
}

const Tooltip = ({ children, overlay, show = false, title }: TooltipProps) => {
  const [showOverlay, setShowOverlay] = useState(false);
  const [referenceElement, setReferenceElement] = useState(null);
  const [popperElement, setPopperElement] = useState(null);
  const [arrowElement, setArrowElement] = useState(null);
  const { styles, attributes } = usePopper(referenceElement, popperElement, {
    modifiers: [{ name: "arrow", options: { element: arrowElement } }],
  });

  const trigger = cloneElement(children, {
    ref: setReferenceElement,
    onMouseEnter: overlay ? () => setShowOverlay(true) : undefined,
    onMouseLeave: overlay ? () => setShowOverlay(false) : undefined,
  });

  return (
    <>
      {trigger}
      {(show || showOverlay) &&
        createPortal(
          <ThemeProvider>
            <ThemeWrapper>
              <div
                // @ts-ignore
                ref={setPopperElement}
                style={styles.popper}
                {...attributes.popper}
                className="z-50 bg-dashboard p-3 border border-divide rounded-md text-sm flex flex-col space-y-2 bg-dashboard-panel"
              >
                <Title title={title} />
                {overlay}
              </div>
            </ThemeWrapper>
          </ThemeProvider>,
          document.body
        )}
    </>
  );
};

const Title = ({ title }) => {
  return <strong className="block break-all">{title}</strong>;
};

export default Tooltip;
